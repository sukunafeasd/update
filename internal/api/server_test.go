package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"universald/internal/db"
	"universald/internal/panel"
)

func TestHandlerSetsSecurityHeadersOnRoot(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!doctype html><html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	srv := &Server{staticDir: staticDir}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Fatalf("expected SAMEORIGIN, got %q", got)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected no-store for root html, got %q", got)
	}
	if got := w.Header().Get("Content-Security-Policy"); !strings.Contains(got, "default-src 'self'") {
		t.Fatalf("expected CSP header, got %q", got)
	}
}

func TestPanelOnlyHealthIsPublicAndStable(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!doctype html><html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	srv := NewPanelServer(staticDir, "1.4.4", true, nil, ServerOptions{
		AppEnv:       "staging",
		PublicOrigin: "https://staging.paineldief.example",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if payload["service"] != "painel-dief" {
		t.Fatalf("expected painel-dief service, got %#v", payload["service"])
	}
	if payload["panelOnly"] != true {
		t.Fatalf("expected panelOnly true, got %#v", payload["panelOnly"])
	}
	if payload["env"] != "staging" {
		t.Fatalf("expected env staging, got %#v", payload["env"])
	}
	if payload["publicOrigin"] != "https://staging.paineldief.example" {
		t.Fatalf("expected public origin in health payload, got %#v", payload["publicOrigin"])
	}
	if _, ok := payload["startedAt"]; !ok {
		t.Fatalf("expected startedAt in health payload")
	}
	if _, ok := payload["uptimeSec"]; !ok {
		t.Fatalf("expected uptimeSec in health payload")
	}
}

func TestPanelOnlyHandlerDoesNotExposeLegacyRoutes(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!doctype html><html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	srv := NewPanelServer(staticDir, "1.4.4", true, nil, ServerOptions{})
	req := httptest.NewRequest(http.MethodGet, "/api/plugins", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for legacy route in panel mode, got %d", w.Code)
	}
}

func TestUploadsHandlerServesDangerousFilesAsDownloadSafeText(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "server-upload.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	uploadsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(uploadsDir, "evil.html"), []byte("<script>alert('x')</script>"), 0o644); err != nil {
		t.Fatalf("write upload: %v", err)
	}

	srv := &Server{panelSvc: panel.NewService(store, uploadsDir)}
	req := httptest.NewRequest(http.MethodGet, "/uploads/evil.html", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("expected text/plain for dangerous upload, got %q", got)
	}
	if got := w.Header().Get("Content-Disposition"); !strings.Contains(strings.ToLower(got), "attachment") {
		t.Fatalf("expected attachment disposition, got %q", got)
	}
}

func TestRequestIDHeaderIsAlwaysPresent(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!doctype html><html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	srv := NewPanelServer(staticDir, "1.4.4", true, nil, ServerOptions{})
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if strings.TrimSpace(w.Header().Get("X-Request-ID")) == "" {
		t.Fatalf("expected X-Request-ID header to be set")
	}
}

func TestOpsSummaryRequiresTokenOutsideLoopback(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-ops.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-ops.db"),
		OpsToken: "ops-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ops/summary", nil)
	req.RemoteAddr = "198.51.100.20:443"
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without ops token, got %d", w.Code)
	}
}

func TestOpsSummaryRequiresTokenInProductionEvenOnLoopback(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-ops-production-loopback.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-ops-production-loopback.db"),
		OpsToken: "ops-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ops/summary", nil)
	req.RemoteAddr = "127.0.0.1:443"
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without token on production loopback, got %d", w.Code)
	}
}

func TestOpsSummaryAllowsLoopbackInNonProduction(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-ops-staging-loopback.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "staging",
		DBPath:   filepath.Join(root, "server-ops-staging-loopback.db"),
		OpsToken: "ops-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ops/summary", nil)
	req.RemoteAddr = "127.0.0.1:443"
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for staging loopback ops summary, got %d", w.Code)
	}
}

func TestOpsSummaryAcceptsBearerToken(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-ops.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-ops.db"),
		OpsToken: "ops-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ops/summary", nil)
	req.RemoteAddr = "198.51.100.20:443"
	req.Header.Set("Authorization", "Bearer ops-secret")
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with ops token, got %d", w.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode ops summary: %v", err)
	}
	if payload["env"] != "production" {
		t.Fatalf("expected production env in ops summary, got %#v", payload["env"])
	}
	summary, ok := payload["summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected summary object, got %T", payload["summary"])
	}
	if summary["users"] == nil {
		t.Fatalf("expected users count in ops summary")
	}
}

func TestOpsExportRequiresTokenOutsideLoopback(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-export.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-export.db"),
		OpsToken: "ops-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ops/export", nil)
	req.RemoteAddr = "198.51.100.20:443"
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without ops token, got %d", w.Code)
	}
}

func TestOpsExportStreamsBackupArchive(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-export.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-export.db"),
		OpsToken: "ops-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/ops/export", nil)
	req.RemoteAddr = "198.51.100.20:443"
	req.Header.Set("Authorization", "Bearer ops-secret")
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "application/zip") {
		t.Fatalf("expected zip content type, got %q", got)
	}
	reader, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var foundManifest bool
	var foundSnapshot bool
	for _, file := range reader.File {
		if file.Name == "backup-manifest.json" {
			foundManifest = true
			handle, openErr := file.Open()
			if openErr != nil {
				t.Fatalf("open manifest: %v", openErr)
			}
			raw, readErr := io.ReadAll(handle)
			_ = handle.Close()
			if readErr != nil {
				t.Fatalf("read manifest: %v", readErr)
			}
			if !strings.Contains(string(raw), "\"service\": \"painel-dief\"") {
				t.Fatalf("unexpected manifest content: %s", string(raw))
			}
		}
		if file.Name == "universald.snapshot.db" {
			foundSnapshot = true
		}
	}
	if !foundManifest {
		t.Fatalf("expected backup manifest in zip")
	}
	if !foundSnapshot {
		t.Fatalf("expected database snapshot in zip")
	}
}

func TestEmbeddedDownloadRouteRequiresAccess(t *testing.T) {
	srv := NewPanelServer("", "1.4.4", true, nil, ServerOptions{DownloadPassword: "segredo-app"})
	req := httptest.NewRequest(http.MethodGet, "/downloads/universalD.exe", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for protected embedded download, got %d", w.Code)
	}
}

func TestOpsImportRestoresPreviousSnapshot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PAINEL_DIEF_OWNER_PASSWORD", "TesteOwner#2026")
	store, err := db.Open(filepath.Join(root, "server-import.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-import.db"),
		OpsToken: "ops-secret",
	})

	exportReq := httptest.NewRequest(http.MethodGet, "/api/ops/export", nil)
	exportReq.RemoteAddr = "198.51.100.20:443"
	exportReq.Header.Set("Authorization", "Bearer ops-secret")
	exportRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusOK {
		t.Fatalf("expected export 200, got %d", exportRec.Code)
	}

	owner, _, err := panelSvc.Login("dief", "TesteOwner#2026")
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}
	if _, err := panelSvc.CreateUser(owner, "restore-alvo", "restore@paineldief.local", "Restore#2026", "Restore Alvo", "member"); err != nil {
		t.Fatalf("create extra user before import: %v", err)
	}
	summaryBefore, err := panelSvc.OpsSummary()
	if err != nil {
		t.Fatalf("ops summary before import: %v", err)
	}
	if summaryBefore.Users < 3 {
		t.Fatalf("expected extra user before import, got users=%d", summaryBefore.Users)
	}

	importReq := httptest.NewRequest(http.MethodPost, "/api/ops/import", bytes.NewReader(exportRec.Body.Bytes()))
	importReq.RemoteAddr = "198.51.100.20:443"
	importReq.Header.Set("Authorization", "Bearer ops-secret")
	importReq.Header.Set("Content-Type", "application/zip")
	importRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(importRec, importReq)
	if importRec.Code != http.StatusOK {
		t.Fatalf("expected import 200, got %d: %s", importRec.Code, importRec.Body.String())
	}

	summaryAfter, err := panelSvc.OpsSummary()
	if err != nil {
		t.Fatalf("ops summary after import: %v", err)
	}
	if summaryAfter.Users >= summaryBefore.Users {
		t.Fatalf("expected import to roll back extra user, before=%d after=%d", summaryBefore.Users, summaryAfter.Users)
	}
	if _, err := store.GetPanelUserByLogin("restore-alvo"); err == nil {
		t.Fatalf("expected imported snapshot to remove extra user")
	}
}

func TestOpsImportRequiresTokenOutsideLoopback(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-import-auth.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-import-auth.db"),
		OpsToken: "ops-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ops/import", bytes.NewReader([]byte("fake zip")))
	req.RemoteAddr = "198.51.100.20:443"
	req.Header.Set("Content-Type", "application/zip")
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without ops token, got %d", w.Code)
	}
}

func TestOpsImportRejectsUnexpectedContentType(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "server-import-type.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	srv := NewPanelServer("", "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:   "production",
		DBPath:   filepath.Join(root, "server-import-type.db"),
		OpsToken: "ops-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ops/import", bytes.NewReader([]byte("{}")))
	req.RemoteAddr = "198.51.100.20:443"
	req.Header.Set("Authorization", "Bearer ops-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 for wrong content type, got %d", w.Code)
	}
}

func TestEmbeddedDownloadRouteUnlocksWithPasswordCookie(t *testing.T) {
	srv := NewPanelServer("", "1.4.4", true, nil, ServerOptions{DownloadPassword: "segredo-app"})

	accessReq := httptest.NewRequest(http.MethodPost, "/api/downloads/universald/access", strings.NewReader(`{"password":"segredo-app"}`))
	accessReq.Header.Set("Content-Type", "application/json")
	accessRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(accessRes, accessReq)

	if accessRes.Code != http.StatusOK {
		t.Fatalf("expected 200 when unlocking app download, got %d", accessRes.Code)
	}
	cookies := accessRes.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected download access cookie to be set")
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/downloads/universalD.exe", nil)
	downloadReq.AddCookie(cookies[0])
	downloadRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(downloadRes, downloadReq)

	if downloadRes.Code != http.StatusOK {
		t.Fatalf("expected 200 for embedded download after unlock, got %d", downloadRes.Code)
	}
	if bodyLen := downloadRes.Body.Len(); bodyLen < 1024 {
		t.Fatalf("expected download body larger than 1KB, got %d bytes", bodyLen)
	}
}
