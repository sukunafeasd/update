package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"universald/internal/db"
	"universald/internal/panel"
)

func TestPanelHTTPLifecycle(t *testing.T) {
	root := t.TempDir()

	store, err := db.Open(filepath.Join(root, "panel-http.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	panelSvc := panel.NewService(store, filepath.Join(root, "uploads"))
	if err := panelSvc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	staticDir := filepath.Join(root, "web")
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		t.Fatalf("mkdir web: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!doctype html><html><body>Painel Dief</body></html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	ts := httptest.NewServer(NewPanelServer(staticDir, "1.4.4", true, panelSvc, ServerOptions{
		AppEnv:       "staging",
		PublicOrigin: "https://staging.paineldief.example",
		DBPath:       filepath.Join(root, "panel-http.db"),
		OpsToken:     "ops-secret",
	}).Handler())
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := ts.Client()
	client.Jar = jar

	health := mustRequestJSON(t, client, http.MethodGet, ts.URL+"/api/health", nil, nil, http.StatusOK)
	if health["service"] != "painel-dief" {
		t.Fatalf("unexpected service name: %#v", health["service"])
	}
	if health["panelOnly"] != true {
		t.Fatalf("expected panelOnly=true, got %#v", health["panelOnly"])
	}
	if health["env"] != "staging" {
		t.Fatalf("expected env staging, got %#v", health["env"])
	}

	ready := mustRequestJSON(t, client, http.MethodGet, ts.URL+"/api/ready", nil, nil, http.StatusOK)
	if ready["ready"] != true {
		t.Fatalf("expected ready=true, got %#v", ready["ready"])
	}

	ops := mustRequestJSON(t, client, http.MethodGet, ts.URL+"/api/ops/summary", nil, nil, http.StatusOK)
	if asMap(t, ops["summary"])["users"] == nil {
		t.Fatalf("expected users count in ops summary")
	}

	login := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/login", map[string]any{
		"login":    "dief",
		"password": "valorant",
	}, nil, http.StatusOK)
	if login["ok"] != true {
		t.Fatalf("expected login ok, got %#v", login["ok"])
	}

	bootstrap := asMap(t, login["bootstrap"])
	viewer := asMap(t, bootstrap["viewer"])
	viewerID := asInt64(t, viewer["id"])
	rooms := asSlice(t, bootstrap["rooms"])
	room := findRoomMapBySlug(t, rooms, "chat-geral")
	roomID := asInt64(t, room["id"])

	profile := mustRequestJSON(t, client, http.MethodGet, fmt.Sprintf("%s/api/panel/social/profile?userId=%d", ts.URL, viewerID), nil, nil, http.StatusOK)
	profileUser := asMap(t, asMap(t, profile["profile"])["user"])
	if asInt64(t, profileUser["userId"]) != viewerID {
		t.Fatalf("unexpected social profile user id")
	}

	updatedProfile := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/profile", map[string]any{
		"displayName": "Dief Teste",
		"bio":         "Teste HTTP do Painel Dief",
		"theme":       "matrix",
		"accentColor": "#33ff88",
		"status":      "online",
		"statusText":  "lapidando o painel",
	}, nil, http.StatusOK)
	if asMap(t, updatedProfile["viewer"])["statusText"] != "lapidando o painel" {
		t.Fatalf("statusText not updated")
	}

	upload := mustUploadFile(t, client, ts.URL+"/api/panel/upload", "files", "audit-http.png", "image/png", []byte{
		0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n',
		0, 0, 0, 0x0d, 'I', 'H', 'D', 'R',
		0, 0, 0, 0x01, 0, 0, 0, 0x01, 0x08, 0x02, 0, 0, 0, 0x90, 0x77, 0x53, 0xde,
		0, 0, 0, 0x0c, 'I', 'D', 'A', 'T', 0x08, 0x99, 0x63, 0xf8, 0xcf, 0xc0, 0, 0, 0x03, 0x01, 0x01, 0, 0x18, 0xdd, 0x8d, 0xb5,
		0, 0, 0, 0, 'I', 'E', 'N', 'D', 0xae, 'B', 0x60, 0x82,
	}, http.StatusOK)
	attachment := asMap(t, upload["attachment"])
	if strings.TrimSpace(asString(t, attachment["url"])) == "" {
		t.Fatalf("expected uploaded attachment url")
	}

	invalidUpload := mustUploadFile(t, client, ts.URL+"/api/panel/upload", "files", "audit-http.html", "text/html", []byte("<html>nao</html>"), http.StatusUnsupportedMediaType)
	if !strings.Contains(strings.ToLower(asString(t, invalidUpload["error"])), "formato") {
		t.Fatalf("expected invalid upload error, got %#v", invalidUpload["error"])
	}

	uploadResp, err := client.Get(ts.URL + asString(t, attachment["url"]))
	if err != nil {
		t.Fatalf("fetch upload back: %v", err)
	}
	defer uploadResp.Body.Close()
	if uploadResp.StatusCode != http.StatusOK {
		t.Fatalf("expected upload fetch 200, got %d", uploadResp.StatusCode)
	}

	posted := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/messages", map[string]any{
		"roomId":     roomID,
		"body":       "mensagem http do painel",
		"kind":       attachment["kind"],
		"attachment": attachment,
		"replyToId":  0,
	}, nil, http.StatusOK)
	message := asMap(t, posted["message"])
	messageID := asInt64(t, message["id"])

	edited := mustRequestJSON(t, client, http.MethodPut, ts.URL+"/api/panel/messages", map[string]any{
		"roomId":    roomID,
		"messageId": messageID,
		"body":      "mensagem http refinada",
	}, nil, http.StatusOK)
	if asMap(t, edited["message"])["body"] != "mensagem http refinada" {
		t.Fatalf("message body not edited")
	}

	reaction := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/reactions/toggle", map[string]any{
		"roomId":    roomID,
		"messageId": messageID,
		"emoji":     "🔥",
	}, nil, http.StatusOK)
	reactionMessage := asMap(t, reaction["message"])
	reactions := asSlice(t, reactionMessage["reactions"])
	if len(reactions) == 0 {
		t.Fatalf("expected reaction toggle to append reactions")
	}

	favorite := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/favorites/toggle", map[string]any{
		"roomId":    roomID,
		"messageId": messageID,
	}, nil, http.StatusOK)
	if favorite["favorited"] != true {
		t.Fatalf("expected favorite toggle to favorite")
	}

	pin := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/pins/toggle", map[string]any{
		"roomId":    roomID,
		"messageId": messageID,
	}, nil, http.StatusOK)
	if pin["pinned"] != true {
		t.Fatalf("expected pin toggle to pin")
	}

	search := mustRequestJSON(t, client, http.MethodGet, ts.URL+"/api/panel/search?query=refinada&limit=10", nil, nil, http.StatusOK)
	if !containsMessageID(t, search["messages"], messageID) {
		t.Fatalf("search did not find edited message")
	}

	poll := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/polls", map[string]any{
		"roomId":   roomID,
		"question": "Qual modo fecha hoje?",
		"options":  []string{"ranked", "casual"},
	}, nil, http.StatusOK)
	pollMap := asMap(t, poll["poll"])
	pollID := asInt64(t, pollMap["id"])
	pollOptions := asSlice(t, pollMap["options"])
	if len(pollOptions) < 1 {
		t.Fatalf("expected poll options")
	}

	vote := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/polls/vote-toggle", map[string]any{
		"roomId":   roomID,
		"pollId":   pollID,
		"optionId": asInt64(t, asMap(t, pollOptions[0])["id"]),
	}, nil, http.StatusOK)
	if vote["voted"] != true {
		t.Fatalf("expected vote toggle to vote")
	}

	event := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/events", map[string]any{
		"title":       "Sessao de teste",
		"description": "Validando o fluxo HTTP completo",
		"roomId":      roomID,
		"startsAt":    time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
	}, nil, http.StatusOK)
	eventMap := asMap(t, event["event"])
	eventID := asInt64(t, eventMap["id"])

	rsvp := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/events/rsvp-toggle", map[string]any{
		"eventId": eventID,
	}, nil, http.StatusOK)
	if rsvp["joined"] != true {
		t.Fatalf("expected event RSVP to join")
	}

	mustRequestJSON(t, client, http.MethodDelete, ts.URL+"/api/panel/polls", map[string]any{
		"roomId": roomID,
		"pollId": pollID,
	}, nil, http.StatusOK)
	pollsAfterDelete := mustRequestJSON(t, client, http.MethodGet, fmt.Sprintf("%s/api/panel/polls?roomId=%d&limit=12", ts.URL, roomID), nil, nil, http.StatusOK)
	if containsEntityID(t, pollsAfterDelete["polls"], pollID) {
		t.Fatalf("deleted poll still listed")
	}

	mustRequestJSON(t, client, http.MethodDelete, ts.URL+"/api/panel/events", map[string]any{
		"eventId": eventID,
	}, nil, http.StatusOK)
	eventsAfterDelete := mustRequestJSON(t, client, http.MethodGet, ts.URL+"/api/panel/events?limit=12", nil, nil, http.StatusOK)
	if containsEntityID(t, eventsAfterDelete["events"], eventID) {
		t.Fatalf("deleted event still listed")
	}

	messages := mustRequestJSON(t, client, http.MethodGet, fmt.Sprintf("%s/api/panel/messages?roomId=%d&limit=20", ts.URL, roomID), nil, nil, http.StatusOK)
	if !containsMessageID(t, messages["messages"], messageID) {
		t.Fatalf("message list did not include posted message")
	}

	mustRequestJSON(t, client, http.MethodDelete, ts.URL+"/api/panel/messages", map[string]any{
		"roomId":    roomID,
		"messageId": messageID,
	}, nil, http.StatusOK)

	memberStamp := time.Now().UTC().UnixNano()
	memberName := fmt.Sprintf("membro%d", memberStamp)
	memberEmail := fmt.Sprintf("membro%d@paineldief.local", memberStamp)
	memberPassword := "Membro#2026"
	mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/users", map[string]any{
		"username":    memberName,
		"email":       memberEmail,
		"password":    memberPassword,
		"displayName": "Membro HTTP",
		"role":        "member",
	}, nil, http.StatusOK)
	mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/logout", nil, nil, http.StatusOK)
	memberLogin := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/login", map[string]any{
		"login":    memberName,
		"password": memberPassword,
	}, nil, http.StatusOK)
	if memberLogin["ok"] != true {
		t.Fatalf("expected member login ok")
	}
	terminalDenied := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/terminal/run", map[string]any{
		"command": "Get-Date",
	}, nil, http.StatusForbidden)
	if !strings.Contains(strings.ToLower(asString(t, terminalDenied["error"])), "admin") {
		t.Fatalf("expected terminal admin error, got %#v", terminalDenied["error"])
	}
	createUserDenied := mustRequestJSON(t, client, http.MethodPost, ts.URL+"/api/panel/users", map[string]any{
		"username":    "nao-pode",
		"email":       "nao-pode@paineldief.local",
		"password":    "Nada#2026",
		"displayName": "Nao Pode",
		"role":        "member",
	}, nil, http.StatusForbidden)
	if !strings.Contains(strings.ToLower(asString(t, createUserDenied["error"])), "dono") {
		t.Fatalf("expected owner restriction error, got %#v", createUserDenied["error"])
	}
}

func mustUploadFile(t *testing.T, client *http.Client, url string, fieldName string, fileName string, contentType string, payload []byte, wantStatus int) map[string]any {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		t.Fatalf("new upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if strings.TrimSpace(contentType) != "" {
		req.Header.Set("X-Upload-Content-Type", contentType)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("perform upload request: %v", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read upload response: %v", err)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected upload status %d: %s", resp.StatusCode, string(raw))
	}

	var payloadMap map[string]any
	if err := json.Unmarshal(raw, &payloadMap); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	return payloadMap
}

func mustRequestJSON(t *testing.T, client *http.Client, method string, url string, payload any, headers map[string]string, wantStatus int) map[string]any {
	t.Helper()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status %d for %s %s: %s", resp.StatusCode, method, url, string(raw))
	}

	if len(raw) == 0 {
		return map[string]any{}
	}

	var payloadMap map[string]any
	if err := json.Unmarshal(raw, &payloadMap); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payloadMap
}

func asMap(t *testing.T, value any) map[string]any {
	t.Helper()
	item, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", value)
	}
	return item
}

func asSlice(t *testing.T, value any) []any {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", value)
	}
	return items
}

func asInt64(t *testing.T, value any) int64 {
	t.Helper()
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		t.Fatalf("expected numeric value, got %T", value)
		return 0
	}
}

func asString(t *testing.T, value any) string {
	t.Helper()
	text, ok := value.(string)
	if !ok {
		t.Fatalf("expected string, got %T", value)
	}
	return text
}

func findRoomMapBySlug(t *testing.T, rooms []any, slug string) map[string]any {
	t.Helper()
	for _, raw := range rooms {
		room := asMap(t, raw)
		if strings.EqualFold(strings.TrimSpace(asString(t, room["slug"])), strings.TrimSpace(slug)) {
			return room
		}
	}
	t.Fatalf("room with slug %q not found", slug)
	return nil
}

func containsMessageID(t *testing.T, value any, messageID int64) bool {
	t.Helper()
	return containsEntityID(t, value, messageID)
}

func containsEntityID(t *testing.T, value any, entityID int64) bool {
	t.Helper()
	items := asSlice(t, value)
	for _, raw := range items {
		item := asMap(t, raw)
		if asInt64(t, item["id"]) == entityID {
			return true
		}
	}
	return false
}
