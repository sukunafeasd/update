package api

import (
	"encoding/json"
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
