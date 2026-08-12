package handler

import (
	"net/http"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/dto"
	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type ClientHandler struct {
	client port.ClientService
}

func NewClientHandler(client port.ClientService) *ClientHandler {
	return &ClientHandler{client: client}
}

func (h *ClientHandler) GetFeatures(w http.ResponseWriter, r *http.Request) {
	tok := middleware.TokenFromContext(r.Context())
	if tok == nil {
		writeError(w, apperr.ErrUnauthorized)
		return
	}
	resp, err := h.client.GetFeatures(r.Context(), tok)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ClientHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	tok := middleware.TokenFromContext(r.Context())
	if tok == nil {
		writeError(w, apperr.ErrUnauthorized)
		return
	}
	var payload port.MetricsPayload
	if err := decodeAndValidate(r, &payload); err != nil {
		writeError(w, err)
		return
	}
	if err := h.client.IngestMetrics(r.Context(), tok, payload); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *ClientHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	tok := middleware.TokenFromContext(r.Context())
	if tok == nil {
		writeError(w, apperr.ErrUnauthorized)
		return
	}
	var req dto.EvaluateRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	enabled, err := h.client.Evaluate(r.Context(), tok, req.FlagKey, req.Context)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}
