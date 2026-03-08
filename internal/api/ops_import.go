package api

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) handleOpsImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
	r.Body = http.MaxBytesReader(w, r.Body, 512<<20)
	if err := s.panelSvc.RestoreBackupArchive(r.Body, "ops-import"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"message":   "Backup importado e restaurado na instancia atual.",
		"version":   s.panelSvc.Version(),
		"timestamp": time.Now().UTC(),
	})
}
