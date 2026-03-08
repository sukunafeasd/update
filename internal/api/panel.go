package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"universald/internal/model"
	"universald/internal/panel"
)

const panelSessionCookie = "painel_dief_session"

type panelLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type panelPresenceRequest struct {
	RoomID int64  `json:"roomId"`
	Status string `json:"status"`
}

type panelEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	RoomID      int64  `json:"roomId"`
	StartsAt    string `json:"startsAt"`
}

type panelEventRSVPRequest struct {
	EventID int64 `json:"eventId"`
}

type panelEventDeleteRequest struct {
	EventID int64 `json:"eventId"`
}

type panelPollRequest struct {
	RoomID   int64    `json:"roomId"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type panelPollVoteRequest struct {
	RoomID   int64 `json:"roomId"`
	PollID   int64 `json:"pollId"`
	OptionID int64 `json:"optionId"`
}

type panelPollDeleteRequest struct {
	RoomID int64 `json:"roomId"`
	PollID int64 `json:"pollId"`
}

type panelMessageRequest struct {
	RoomID     int64                  `json:"roomId"`
	Body       string                 `json:"body"`
	Kind       string                 `json:"kind"`
	ReplyToID  int64                  `json:"replyToId"`
	Attachment *model.PanelAttachment `json:"attachment"`
}

type panelMessageUpdateRequest struct {
	RoomID    int64  `json:"roomId"`
	MessageID int64  `json:"messageId"`
	Body      string `json:"body"`
}

type panelMessageActionRequest struct {
	RoomID    int64 `json:"roomId"`
	MessageID int64 `json:"messageId"`
}

type panelUnlockRequest struct {
	RoomID   int64  `json:"roomId"`
	Password string `json:"password"`
}

type panelCreateUserRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type panelProfileRequest struct {
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	Theme       string `json:"theme"`
	AccentColor string `json:"accentColor"`
	AvatarURL   string `json:"avatarUrl"`
	Status      string `json:"status"`
	StatusText  string `json:"statusText"`
}

type panelAIRequest struct {
	RoomID int64  `json:"roomId"`
	Prompt string `json:"prompt"`
}

type panelTerminalRequest struct {
	Command string `json:"command"`
}

type panelTypingRequest struct {
	RoomID int64 `json:"roomId"`
	Active bool  `json:"active"`
}

type panelReactionRequest struct {
	RoomID    int64  `json:"roomId"`
	MessageID int64  `json:"messageId"`
	Emoji     string `json:"emoji"`
}

type panelSearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type panelDMRequest struct {
	TargetUserID int64 `json:"targetUserId"`
}

type panelSocialUserActionRequest struct {
	TargetUserID int64 `json:"targetUserId"`
}

func (s *Server) handlePanelLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.panelSvc == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("painel indisponivel"))
		return
	}

	var req panelLoginRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	user, session, err := s.panelSvc.Login(req.Login, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, panel.ErrLoginRateLimited) {
			status = http.StatusTooManyRequests
			var retryAfter interface{ RetryAfter() time.Duration }
			if errors.As(err, &retryAfter) {
				seconds := int(retryAfter.RetryAfter().Seconds())
				if seconds < 1 {
					seconds = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(seconds))
			}
		}
		writeError(w, status, err)
		return
	}
	writePanelCookie(w, r, session)
	bootstrap, err := s.panelSvc.Bootstrap(user, session.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"message":   "Acesso liberado ao Painel Dief.",
		"sessionId": session.ID,
		"bootstrap": bootstrap,
	})
}

func (s *Server) handlePanelLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.panelSvc == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("painel indisponivel"))
		return
	}
	sessionID := readPanelSessionID(r)
	if sessionID != "" {
		_ = s.panelSvc.Logout(sessionID)
	}
	clearPanelCookie(w, r)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handlePanelBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	bootstrap, err := s.panelSvc.Bootstrap(user, sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	bootstrap.SessionID = sessionID
	writeJSON(w, http.StatusOK, bootstrap)
}

func (s *Server) handlePanelPresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelPresenceRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.panelSvc.UpdatePresence(user, req.RoomID, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "serverTime": time.Now().UTC()})
}

func (s *Server) handlePanelMessages(w http.ResponseWriter, r *http.Request) {
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		roomID, err := parseInt64Required(r.URL.Query().Get("roomId"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		items, room, err := s.panelSvc.ListMessages(user, sessionID, roomID, limit)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(strings.ToLower(err.Error()), "acesso") {
				status = http.StatusForbidden
			}
			writeError(w, status, err)
			return
		}
		pins, err := s.panelSvc.ListPinnedMessages(user, sessionID, roomID, 6)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(strings.ToLower(err.Error()), "acesso") {
				status = http.StatusForbidden
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"room": room, "messages": items, "pins": pins, "version": s.panelSvc.Version()})
	case http.MethodPost:
		var req panelMessageRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		msg, err := s.panelSvc.PostMessage(user, sessionID, req.RoomID, req.Body, req.Kind, req.Attachment, req.ReplyToID)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(strings.ToLower(err.Error()), "acesso") {
				status = http.StatusForbidden
			} else if strings.Contains(strings.ToLower(err.Error()), "vazia") || strings.Contains(strings.ToLower(err.Error()), "spam") || strings.Contains(strings.ToLower(err.Error()), "responder") {
				status = http.StatusBadRequest
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": msg})
	case http.MethodPut:
		var req panelMessageUpdateRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		msg, err := s.panelSvc.EditMessage(user, sessionID, req.RoomID, req.MessageID, req.Body)
		if err != nil {
			status := http.StatusBadRequest
			if strings.Contains(strings.ToLower(err.Error()), "acesso") || strings.Contains(strings.ToLower(err.Error()), "autor") || strings.Contains(strings.ToLower(err.Error()), "owner") || strings.Contains(strings.ToLower(err.Error()), "admin") {
				status = http.StatusForbidden
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": msg})
	case http.MethodDelete:
		var req panelMessageActionRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := s.panelSvc.DeleteMessage(user, sessionID, req.RoomID, req.MessageID); err != nil {
			status := http.StatusBadRequest
			if strings.Contains(strings.ToLower(err.Error()), "acesso") || strings.Contains(strings.ToLower(err.Error()), "autor") || strings.Contains(strings.ToLower(err.Error()), "owner") || strings.Contains(strings.ToLower(err.Error()), "admin") {
				status = http.StatusForbidden
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "messageId": req.MessageID})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePanelUnlockRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelUnlockRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.panelSvc.UnlockRoom(user, sessionID, req.RoomID, req.Password); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "Sala destrancada."})
}

func (s *Server) handlePanelUsers(w http.ResponseWriter, r *http.Request) {
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		users, err := s.panelSvc.ListUsers(user)
		if err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"users": users})
	case http.MethodPost:
		var req panelCreateUserRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		created, err := s.panelSvc.CreateUser(user, req.Username, req.Email, req.Password, req.DisplayName, req.Role)
		if err != nil {
			status := http.StatusBadRequest
			if strings.Contains(strings.ToLower(err.Error()), "dono") {
				status = http.StatusForbidden
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": created})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePanelProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelProfileRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	updated, err := s.panelSvc.UpdateProfile(user, req.DisplayName, req.Bio, req.Theme, req.AccentColor, req.AvatarURL, req.Status, req.StatusText)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "viewer": updated})
}

func (s *Server) handlePanelUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	limit := s.panelSvc.UploadLimitForRole(user.Role) + (2 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			status = http.StatusRequestEntityTooLarge
		}
		writeError(w, status, fmt.Errorf("upload invalido: %w", err))
		return
	}
	headers := make([]*multipart.FileHeader, 0)
	for _, items := range r.MultipartForm.File {
		headers = append(headers, items...)
	}
	if len(headers) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("nenhum arquivo veio no upload"))
		return
	}
	attachments := make([]model.PanelAttachment, 0, len(headers))
	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		contentType := header.Header.Get("Content-Type")
		if strings.TrimSpace(contentType) == "" {
			contentType = "application/octet-stream"
		}
		attachment, saveErr := s.panelSvc.SaveUpload(user, header.Filename, contentType, file)
		_ = file.Close()
		if saveErr != nil {
			status := http.StatusBadRequest
			lower := strings.ToLower(saveErr.Error())
			if strings.Contains(lower, "limite") {
				status = http.StatusRequestEntityTooLarge
			} else if strings.Contains(lower, "formato nao permitido") {
				status = http.StatusUnsupportedMediaType
			}
			writeError(w, status, saveErr)
			return
		}
		attachments = append(attachments, attachment)
	}
	if len(attachments) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("upload vazio"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"attachment":  attachments[0],
		"attachments": attachments,
	})
}

func (s *Server) handlePanelAIChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelAIRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userMsg, reply, err := s.panelSvc.PostAIExchange(user, sessionID, req.RoomID, req.Prompt)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "question": userMsg, "reply": reply})
}

func (s *Server) handlePanelTerminalRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelTerminalRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.panelSvc.RunTerminal(user, req.Command)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "admin") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result})
}

func (s *Server) handlePanelTyping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelTypingRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.panelSvc.UpdateTyping(user, sessionID, req.RoomID, req.Active); err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "acesso") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handlePanelReactionToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelReactionRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	msg, err := s.panelSvc.ToggleReaction(user, sessionID, req.RoomID, req.MessageID, req.Emoji)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "acesso") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": msg})
}

func (s *Server) handlePanelPinToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelMessageActionRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	msg, pinned, err := s.panelSvc.TogglePin(user, sessionID, req.RoomID, req.MessageID)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "acesso") || strings.Contains(strings.ToLower(err.Error()), "autor") || strings.Contains(strings.ToLower(err.Error()), "owner") || strings.Contains(strings.ToLower(err.Error()), "admin") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": msg, "pinned": pinned})
}

func (s *Server) handlePanelFavoriteToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelMessageActionRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	msg, favorited, err := s.panelSvc.ToggleFavorite(user, sessionID, req.RoomID, req.MessageID)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "acesso") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": msg, "favorited": favorited})
}

func (s *Server) handlePanelSearch(w http.ResponseWriter, r *http.Request) {
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		query := strings.TrimSpace(r.URL.Query().Get("query"))
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		result, err := s.panelSvc.Search(user, sessionID, query, limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	case http.MethodPost:
		var req panelSearchRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		result, err := s.panelSvc.Search(user, sessionID, req.Query, req.Limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePanelLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
	items, err := s.panelSvc.ListLogs(user, limit)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "restritos") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"logs": items})
}

func (s *Server) handlePanelEvents(w http.ResponseWriter, r *http.Request) {
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		items, err := s.panelSvc.ListEvents(user, sessionID, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": items})
	case http.MethodPost:
		var req panelEventRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		startsAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.StartsAt))
		if err != nil {
			if alt, altErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(req.StartsAt)); altErr == nil {
				startsAt = alt
			} else {
				writeError(w, http.StatusBadRequest, errors.New("data do evento invalida"))
				return
			}
		}
		item, err := s.panelSvc.CreateEvent(user, sessionID, req.Title, req.Description, req.RoomID, startsAt)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(strings.ToLower(err.Error()), "inval") || strings.Contains(strings.ToLower(err.Error()), "curto") || strings.Contains(strings.ToLower(err.Error()), "limite") || strings.Contains(strings.ToLower(err.Error()), "velho") || strings.Contains(strings.ToLower(err.Error()), "pode") {
				status = http.StatusBadRequest
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "event": item, "version": s.panelSvc.Version()})
	case http.MethodDelete:
		var req panelEventDeleteRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := s.panelSvc.DeleteEvent(user, sessionID, req.EventID); err != nil {
			status := http.StatusBadRequest
			lower := strings.ToLower(err.Error())
			if strings.Contains(lower, "so criador") {
				status = http.StatusForbidden
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "eventId": req.EventID, "version": s.panelSvc.Version()})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePanelPolls(w http.ResponseWriter, r *http.Request) {
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		roomID, err := parseInt64Required(r.URL.Query().Get("roomId"))
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("roomId obrigatorio"))
			return
		}
		limit := 12
		if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
			if parsed, parseErr := strconv.Atoi(rawLimit); parseErr == nil && parsed > 0 {
				limit = parsed
			}
		}
		items, err := s.panelSvc.ListPolls(user, sessionID, roomID, limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"polls": items})
	case http.MethodPost:
		var req panelPollRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := s.panelSvc.CreatePoll(user, sessionID, req.RoomID, req.Question, req.Options)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "poll": item, "version": s.panelSvc.Version()})
	case http.MethodDelete:
		var req panelPollDeleteRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := s.panelSvc.DeletePoll(user, sessionID, req.RoomID, req.PollID); err != nil {
			status := http.StatusBadRequest
			lower := strings.ToLower(err.Error())
			if strings.Contains(lower, "so criador") {
				status = http.StatusForbidden
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "pollId": req.PollID, "version": s.panelSvc.Version()})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePanelPollVoteToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelPollVoteRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, voted, err := s.panelSvc.VotePoll(user, sessionID, req.RoomID, req.PollID, req.OptionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "poll": item, "voted": voted, "version": s.panelSvc.Version()})
}

func (s *Server) handlePanelEventRSVPToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelEventRSVPRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, joined, err := s.panelSvc.ToggleEventRSVP(user, sessionID, req.EventID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "evento") || strings.Contains(strings.ToLower(err.Error()), "alcance") {
			status = http.StatusBadRequest
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "event": item, "joined": joined, "version": s.panelSvc.Version()})
}

func (s *Server) handlePanelDMOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelDMRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	room, err := s.panelSvc.OpenDirectRoom(user, req.TargetUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "room": room, "version": s.panelSvc.Version()})
}

func (s *Server) handlePanelSocialProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	targetUserID, err := parseInt64Required(r.URL.Query().Get("userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	profile, err := s.panelSvc.GetSocialProfile(user, targetUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": profile})
}

func (s *Server) handlePanelSocialBlockToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelSocialUserActionRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	blocked, err := s.panelSvc.ToggleBlock(user, req.TargetUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "blocked": blocked})
}

func (s *Server) handlePanelSocialMuteToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelSocialUserActionRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	muted, err := s.panelSvc.ToggleMute(user, req.TargetUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "muted": muted})
}

func (s *Server) handlePanelStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, sessionID, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, errors.New("streaming unsupported"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	send := func() bool {
		payload, err := s.panelSvc.StreamSnapshot(user, sessionID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return false
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return false
		}
		_, _ = fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", string(body))
		flusher.Flush()
		return true
	}
	if !send() {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if !send() {
				return
			}
		}
	}
}

func (s *Server) requirePanelAuth(w http.ResponseWriter, r *http.Request) (model.PanelUser, model.PanelSession, string, bool) {
	if s.panelSvc == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("painel indisponivel"))
		return model.PanelUser{}, model.PanelSession{}, "", false
	}
	sessionID := readPanelSessionID(r)
	if strings.TrimSpace(sessionID) == "" {
		writeError(w, http.StatusUnauthorized, errors.New("login necessario"))
		return model.PanelUser{}, model.PanelSession{}, "", false
	}
	user, session, err := s.panelSvc.Authenticate(sessionID)
	if err != nil {
		clearPanelCookie(w, r)
		writeError(w, http.StatusUnauthorized, err)
		return model.PanelUser{}, model.PanelSession{}, "", false
	}
	return user, session, sessionID, true
}

func writePanelCookie(w http.ResponseWriter, r *http.Request, session model.PanelSession) {
	http.SetCookie(w, &http.Cookie{
		Name:     panelSessionCookie,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		Expires:  session.ExpiresAt,
	})
}

func clearPanelCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     panelSessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func readPanelSessionID(r *http.Request) string {
	if cookie, err := r.Cookie(panelSessionCookie); err == nil && cookie != nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value)
	}
	return strings.TrimSpace(r.Header.Get("X-Panel-Session"))
}

func parseInt64Required(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("parametro obrigatorio ausente")
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parametro invalido: %w", err)
	}
	return value, nil
}
