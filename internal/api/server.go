package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"universald/internal/panel"
	webassets "universald/web"
)

type ServerOptions struct {
	AppEnv           string
	PublicOrigin     string
	DBPath           string
	OpsToken         string
	DownloadPassword string
}

type Server struct {
	panelSvc     *panel.Service
	staticDir    string
	version      string
	service      string
	appEnv       string
	publicOrigin string
	dbPath       string
	opsToken     string
	downloadPass string
	safeMode     bool
	panelOnly    bool
	startedAt    time.Time
}

func NewPanelServer(staticDir, version string, safeMode bool, panelSvc *panel.Service, opts ServerOptions) *Server {
	return &Server{
		panelSvc:     panelSvc,
		staticDir:    staticDir,
		version:      version,
		service:      "painel-dief",
		appEnv:       normalizeAppEnv(opts.AppEnv),
		publicOrigin: strings.TrimRight(strings.TrimSpace(opts.PublicOrigin), "/"),
		dbPath:       strings.TrimSpace(opts.DBPath),
		opsToken:     strings.TrimSpace(opts.OpsToken),
		downloadPass: strings.TrimSpace(opts.DownloadPassword),
		safeMode:     safeMode,
		panelOnly:    true,
		startedAt:    time.Now().UTC(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/ready", s.handleReady)
	mux.HandleFunc("/api/ops/summary", s.handleOpsSummary)
	mux.HandleFunc("/api/ops/export", s.handleOpsExport)
	mux.HandleFunc("/api/downloads/universald/access", s.handleUniversalDDownloadAccess)
	mux.HandleFunc("/downloads/universalD.exe", s.handleUniversalDDownload)
	s.registerPanelRoutes(mux)
	if s.panelSvc != nil && strings.TrimSpace(s.panelSvc.UploadsDir()) != "" {
		mux.Handle("/uploads/", s.uploadsHandler())
	}
	mux.Handle("/", s.staticHandler())
	return withRequestLog(withRequestID(withSecurityHeaders(mux)))
}

func (s *Server) registerPanelRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/panel/login", s.handlePanelLogin)
	mux.HandleFunc("/api/panel/logout", s.handlePanelLogout)
	mux.HandleFunc("/api/panel/bootstrap", s.handlePanelBootstrap)
	mux.HandleFunc("/api/panel/presence", s.handlePanelPresence)
	mux.HandleFunc("/api/panel/messages", s.handlePanelMessages)
	mux.HandleFunc("/api/panel/rooms/unlock", s.handlePanelUnlockRoom)
	mux.HandleFunc("/api/panel/users", s.handlePanelUsers)
	mux.HandleFunc("/api/panel/profile", s.handlePanelProfile)
	mux.HandleFunc("/api/panel/upload", s.handlePanelUpload)
	mux.HandleFunc("/api/panel/ai/chat", s.handlePanelAIChat)
	mux.HandleFunc("/api/panel/terminal/run", s.handlePanelTerminalRun)
	mux.HandleFunc("/api/panel/typing", s.handlePanelTyping)
	mux.HandleFunc("/api/panel/reactions/toggle", s.handlePanelReactionToggle)
	mux.HandleFunc("/api/panel/pins/toggle", s.handlePanelPinToggle)
	mux.HandleFunc("/api/panel/favorites/toggle", s.handlePanelFavoriteToggle)
	mux.HandleFunc("/api/panel/search", s.handlePanelSearch)
	mux.HandleFunc("/api/panel/logs", s.handlePanelLogs)
	mux.HandleFunc("/api/panel/events", s.handlePanelEvents)
	mux.HandleFunc("/api/panel/events/rsvp-toggle", s.handlePanelEventRSVPToggle)
	mux.HandleFunc("/api/panel/polls", s.handlePanelPolls)
	mux.HandleFunc("/api/panel/polls/vote-toggle", s.handlePanelPollVoteToggle)
	mux.HandleFunc("/api/panel/dms/open", s.handlePanelDMOpen)
	mux.HandleFunc("/api/panel/social/profile", s.handlePanelSocialProfile)
	mux.HandleFunc("/api/panel/social/block-toggle", s.handlePanelSocialBlockToggle)
	mux.HandleFunc("/api/panel/social/mute-toggle", s.handlePanelSocialMuteToggle)
	mux.HandleFunc("/api/panel/stream", s.handlePanelStream)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	service := strings.TrimSpace(s.service)
	if service == "" {
		service = "painel-dief"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"service":      service,
		"version":      s.version,
		"env":          normalizeAppEnv(s.appEnv),
		"publicOrigin": s.publicOrigin,
		"safeMode":     s.safeMode,
		"authToken":    false,
		"opsSecured":   strings.TrimSpace(s.opsToken) != "",
		"panelOnly":    true,
		"streaming":    true,
		"startedAt":    s.startedAt,
		"uptimeSec":    int64(time.Since(s.startedAt).Seconds()),
		"timestamp":    time.Now().UTC(),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.panelSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":    "degraded",
			"ready":     false,
			"service":   s.service,
			"version":   s.version,
			"env":       normalizeAppEnv(s.appEnv),
			"timestamp": time.Now().UTC(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ready",
		"ready":     true,
		"service":   s.service,
		"version":   s.version,
		"env":       normalizeAppEnv(s.appEnv),
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleOpsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.panelSvc == nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("painel indisponivel"))
		return
	}
	if !s.authorizeOpsRequest(r) {
		writeError(w, http.StatusForbidden, fmt.Errorf("acesso ops negado"))
		return
	}
	summary, err := s.panelSvc.OpsSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service":      s.service,
		"version":      s.version,
		"env":          normalizeAppEnv(s.appEnv),
		"publicOrigin": s.publicOrigin,
		"safeMode":     s.safeMode,
		"startedAt":    s.startedAt,
		"uptimeSec":    int64(time.Since(s.startedAt).Seconds()),
		"dbFileBytes":  fileSizeBytes(s.dbPath),
		"summary":      summary,
		"timestamp":    time.Now().UTC(),
	})
}

func (s *Server) handleOpsExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.panelSvc == nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("painel indisponivel"))
		return
	}
	if !s.authorizeOpsRequest(r) {
		writeError(w, http.StatusForbidden, fmt.Errorf("acesso ops negado"))
		return
	}

	fileName := "painel-dief-export-" + time.Now().UTC().Format("20060102-150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))
	if _, err := s.panelSvc.WriteBackupArchive(w, "ops-export"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
}

func (s *Server) staticHandler() http.Handler {
	if s.staticDir != "" {
		if _, err := os.Stat(filepath.Join(s.staticDir, "index.html")); err == nil {
			return http.FileServer(http.Dir(s.staticDir))
		}
	}
	return http.FileServer(http.FS(webassets.Files()))
}

func (s *Server) handleUniversalDDownloadAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if strings.TrimSpace(s.downloadPass) == "" {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("download privado nao configurado"))
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(req.Password)), []byte(s.downloadPass)) != 1 {
		writeError(w, http.StatusUnauthorized, fmt.Errorf("senha do app incorreta"))
		return
	}
	writeDownloadCookie(w, r, s.downloadPass)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"message":     "UniversalD liberado. Pode baixar sem drama.",
		"downloadUrl": "/downloads/universalD.exe",
	})
}

func (s *Server) handleUniversalDDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.canAccessProtectedDownload(r) {
		writeError(w, http.StatusUnauthorized, fmt.Errorf("download privado bloqueado"))
		return
	}
	payload, err := fs.ReadFile(webassets.Files(), "downloads/universalD.exe")
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("app indisponivel"))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": "universalD.exe"}))
	http.ServeContent(w, r, "universalD.exe", time.Time{}, bytes.NewReader(payload))
}

func (s *Server) canAccessProtectedDownload(r *http.Request) bool {
	if s.hasValidPanelSession(r) {
		return true
	}
	if strings.TrimSpace(s.downloadPass) == "" {
		return false
	}
	expected := downloadCookieValue(s.downloadPass)
	if expected == "" {
		return false
	}
	current := readDownloadCookie(r)
	if current == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(current), []byte(expected)) == 1
}

func (s *Server) hasValidPanelSession(r *http.Request) bool {
	if s.panelSvc == nil || r == nil {
		return false
	}
	sessionID := readPanelSessionID(r)
	if strings.TrimSpace(sessionID) == "" {
		return false
	}
	_, _, err := s.panelSvc.Authenticate(sessionID)
	return err == nil
}

func downloadCookieValue(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("painel-dief-download|" + secret))
	return hex.EncodeToString(sum[:])
}

func writeDownloadCookie(w http.ResponseWriter, r *http.Request, secret string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "painel_dief_download",
		Value:    downloadCookieValue(secret),
		Path:     "/downloads/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		Expires:  time.Now().UTC().Add(12 * time.Hour),
	})
}

func readDownloadCookie(r *http.Request) string {
	if cookie, err := r.Cookie("painel_dief_download"); err == nil && cookie != nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}

func decodeJSON(body io.ReadCloser, target any) error {
	defer body.Close()
	if err := json.NewDecoder(io.LimitReader(body, 1<<20)).Decode(target); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, err error) {
	payload := map[string]any{"error": err.Error()}
	if requestID := requestIDFromResponse(w); requestID != "-" {
		payload["requestId"] = requestID
	}
	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode: %v", err)
	}
}

func isSecureRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

func withSecurityHeaders(next http.Handler) http.Handler {
	csp := strings.Join([]string{
		"default-src 'self'",
		"base-uri 'self'",
		"object-src 'none'",
		"form-action 'self'",
		"frame-ancestors 'self'",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"font-src 'self' https://fonts.gstatic.com data:",
		"img-src 'self' data: blob: https:",
		"media-src 'self' data: blob:",
		"connect-src 'self'",
		"frame-src 'self'",
		"manifest-src 'self'",
	}, "; ")
	permissions := "accelerometer=(), autoplay=(self), camera=(), display-capture=(), fullscreen=(self), geolocation=(), microphone=(), payment=(), usb=()"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Content-Security-Policy", csp)
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("Permissions-Policy", permissions)
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "SAMEORIGIN")
		headers.Set("Cross-Origin-Opener-Policy", "same-origin")
		headers.Set("Cross-Origin-Resource-Policy", "same-origin")
		if isSecureRequest(r) {
			headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		path := strings.ToLower(strings.TrimSpace(r.URL.Path))
		switch {
		case strings.HasPrefix(path, "/api/"):
			headers.Set("Cache-Control", "no-store")
			headers.Set("Pragma", "no-cache")
		case path == "/" || strings.HasSuffix(path, ".html"):
			headers.Set("Cache-Control", "no-store")
		case strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js"):
			headers.Set("Cache-Control", "no-cache")
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) uploadsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.panelSvc == nil || strings.TrimSpace(s.panelSvc.UploadsDir()) == "" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rawName := strings.TrimPrefix(strings.TrimSpace(r.URL.Path), "/uploads/")
		rawName = filepath.ToSlash(rawName)
		if rawName == "" || rawName == "." || rawName == "/" || strings.Contains(rawName, "..") || strings.Contains(rawName, "/") {
			http.NotFound(w, r)
			return
		}
		name := filepath.Base(rawName)
		if name == "." || name == "" {
			http.NotFound(w, r)
			return
		}

		fullPath := filepath.Join(s.panelSvc.UploadsDir(), name)
		file, err := os.Open(fullPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		head := make([]byte, 512)
		n, _ := file.Read(head)
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			http.NotFound(w, r)
			return
		}

		contentType, disposition := uploadResponseMeta(name, head[:n])
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", disposition)
		w.Header().Set("Accept-Ranges", "bytes")
		http.ServeContent(w, r, name, info.ModTime(), file)
	})
}

func uploadResponseMeta(name string, head []byte) (string, string) {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(name)))
	quoted := mime.FormatMediaType("attachment", map[string]string{"filename": name})
	inline := mime.FormatMediaType("inline", map[string]string{"filename": name})

	switch ext {
	case ".png":
		return "image/png", inline
	case ".jpg", ".jpeg":
		return "image/jpeg", inline
	case ".gif":
		return "image/gif", inline
	case ".webp":
		return "image/webp", inline
	case ".mp4", ".m4v":
		return "video/mp4", inline
	case ".webm":
		return "video/webm", inline
	case ".mov":
		return "video/quicktime", inline
	case ".mp3":
		return "audio/mpeg", inline
	case ".wav":
		return "audio/wav", inline
	case ".ogg":
		return "audio/ogg", inline
	case ".m4a", ".aac":
		return "audio/mp4", inline
	case ".pdf":
		return "application/pdf", inline
	case ".txt", ".md", ".log", ".go", ".py", ".ts", ".tsx", ".jsx", ".sql", ".yaml", ".yml", ".toml", ".ini", ".csv", ".json":
		return "text/plain; charset=utf-8", inline
	case ".html", ".htm", ".css", ".js", ".mjs", ".svg", ".xml", ".xhtml":
		return "text/plain; charset=utf-8", quoted
	default:
		contentType := strings.TrimSpace(http.DetectContentType(head))
		if contentType == "" || contentType == "application/octet-stream" {
			if byExt := mime.TypeByExtension(ext); byExt != "" {
				contentType = byExt
			} else {
				contentType = "application/octet-stream"
			}
		}
		return contentType, quoted
	}
}

func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s [%s] (%s)", r.Method, r.URL.Path, requestIDFromResponse(w), time.Since(start).Round(time.Millisecond))
	})
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func requestIDFromResponse(w http.ResponseWriter) string {
	if w == nil {
		return "-"
	}
	id := strings.TrimSpace(w.Header().Get("X-Request-ID"))
	if id == "" {
		return "-"
	}
	return id
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("req-%d", time.Now().UTC().UnixNano())
}

func normalizeAppEnv(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "production":
		return "production"
	case "staging":
		return "staging"
	default:
		return "development"
	}
}

func (s *Server) authorizeOpsRequest(r *http.Request) bool {
	if isLoopbackRequest(r) {
		return true
	}
	token := strings.TrimSpace(s.opsToken)
	if token == "" {
		return false
	}
	bearer := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(bearer), "bearer ") {
		return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(bearer[7:])), []byte(token)) == 1
	}
	headerToken := strings.TrimSpace(r.Header.Get("X-Ops-Token"))
	if headerToken != "" {
		return subtle.ConstantTimeCompare([]byte(headerToken), []byte(token)) == 1
	}
	return false
}

func isLoopbackRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	host := strings.TrimSpace(r.RemoteAddr)
	if host == "" {
		return false
	}
	if strings.HasPrefix(host, "[") || strings.Count(host, ":") >= 1 {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			host = parsedHost
		}
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

func fileSizeBytes(path string) int64 {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}
