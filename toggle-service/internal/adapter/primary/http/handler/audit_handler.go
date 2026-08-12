package handler

import (
	"net/http"
	"time"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type AuditHandler struct {
	audit port.AuditService
}

func NewAuditHandler(audit port.AuditService) *AuditHandler {
	return &AuditHandler{audit: audit}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	entityType := r.URL.Query().Get("entityType")

	var since, until *time.Time
	if v := r.URL.Query().Get("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			since = &t
		}
	}
	if v := r.URL.Query().Get("until"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			until = &t
		}
	}

	events, err := h.audit.List(r.Context(), projectID, entityType, since, until)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}
