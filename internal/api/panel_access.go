package api

import (
	"errors"
	"net/http"
	"strings"

	"universald/internal/model"
)

type panelUserRoleRequest struct {
	TargetUserID int64  `json:"targetUserId"`
	Role         string `json:"role"`
}

type panelUserExpelRequest struct {
	TargetUserID int64 `json:"targetUserId"`
}

type panelJoinRequestCreateRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Note        string `json:"note"`
}

type panelJoinRequestCompleteRequest struct {
	Email       string `json:"email"`
	AccessCode  string `json:"accessCode"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

type panelJoinRequestReviewRequest struct {
	RequestID  int64  `json:"requestId"`
	Approve    bool   `json:"approve"`
	ReviewNote string `json:"reviewNote"`
}

func (s *Server) handlePanelUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelUserRoleRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	updated, err := s.panelSvc.UpdateUserRole(user, req.TargetUserID, req.Role)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "dono") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": updated})
}

func (s *Server) handlePanelUserExpel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelUserExpelRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.panelSvc.ExpelUser(user, req.TargetUserID); err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "dono") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "targetUserId": req.TargetUserID})
}

func (s *Server) handlePanelJoinRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.panelSvc == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("painel indisponivel"))
		return
	}
	var req panelJoinRequestCreateRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := s.panelSvc.RequestJoinAccess(req.Email, req.DisplayName, req.Note)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Pedido enviado. Agora o dono decide se tu entra ou toma um nao debochado.",
		"request": item,
	})
}

func (s *Server) handlePanelJoinRequestComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.panelSvc == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("painel indisponivel"))
		return
	}
	var req panelJoinRequestCompleteRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	user, err := s.panelSvc.CompleteJoinAccess(req.Email, req.AccessCode, req.Username, req.DisplayName, req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Bah, cadastro concluido. Agora entra com teu login novo.",
		"user":    user,
	})
}

func (s *Server) handlePanelJoinRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	items, err := s.panelSvc.ListJoinRequests(user)
	if err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": items})
}

func (s *Server) handlePanelJoinRequestReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _, _, ok := s.requirePanelAuth(w, r)
	if !ok {
		return
	}
	var req panelJoinRequestReviewRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := s.panelSvc.ReviewJoinRequest(user, req.RequestID, req.Approve, req.ReviewNote)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "dono") {
			status = http.StatusForbidden
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"request": item,
		"message": reviewToastMessage(item),
	})
}

func reviewToastMessage(item model.PanelJoinRequest) string {
	if strings.TrimSpace(strings.ToLower(item.Status)) == "approved" {
		if item.EmailSent {
			return "Pedido aprovado e codigo enviado por email."
		}
		return "Pedido aprovado e codigo gerado."
	}
	return "Pedido revisado pelo dono."
}
