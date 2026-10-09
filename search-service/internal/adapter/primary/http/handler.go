package http

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/application"
	"github.com/JIeeiroSst/search-service/internal/domain"
	"github.com/JIeeiroSst/search-service/internal/port"
)

const healthTimeout = 2 * time.Second

type Handler struct {
	documents *application.DocumentService
	schema    *application.SchemaService
	health    port.HealthChecker
	key       []byte
}

func NewHandler(cfg *config.Config, documents *application.DocumentService, schema *application.SchemaService, health port.HealthChecker) *Handler {
	return &Handler{documents: documents, schema: schema, health: health, key: []byte(cfg.Secret.AuthorizeKey)}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()
	if err := h.health.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": config.Version})
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.validToken(r.Header.Get("Authorization")) {
			writeError(w, r, domain.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) validToken(header string) bool {
	header = strings.TrimSpace(header)
	if header == "" || len(h.key) == 0 {
		return false
	}
	var token []byte
	if bearer, ok := strings.CutPrefix(header, "Bearer "); ok {
		token = []byte(strings.TrimSpace(bearer))
	} else {
		decoded, err := base64.StdEncoding.DecodeString(header)
		if err != nil {
			return false
		}
		token = decoded
	}
	return subtle.ConstantTimeCompare(token, h.key) == 1
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	msg := "internal server error"
	switch {
	case domain.IsInvalid(err):
		status, msg = http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrNotFound):
		status, msg = http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status, msg = http.StatusUnauthorized, err.Error()
	case errors.Is(err, domain.ErrUnavailable):
		status, msg = http.StatusServiceUnavailable, domain.ErrUnavailable.Error()
	case errors.Is(err, context.DeadlineExceeded):
		status, msg = http.StatusGatewayTimeout, "request timed out"
	}
	if status >= http.StatusInternalServerError {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}
