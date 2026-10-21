package httpapi

import (
	"context"
	"net/http"
	"time"
)

const readyTimeout = 2 * time.Second

type healthHandler struct {
	ready func(ctx context.Context) error
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *healthHandler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (h *healthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.ready != nil {
		ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
		defer cancel()
		if err := h.ready(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
			return
		}
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
