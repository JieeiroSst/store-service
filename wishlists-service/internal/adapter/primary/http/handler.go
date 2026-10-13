package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/config"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"github.com/JIeeiroSst/wishlists-service/pkg/authn"
	"go.uber.org/fx"
)

const maxBodyBytes = 1 << 20

type Handler struct {
	wishlists port.WishlistUsecase
	health    []port.HealthChecker
	limiter   port.RateLimiter
	auth      *authn.Authenticator
	writes    int
}

type Params struct {
	fx.In

	Wishlists port.WishlistUsecase
	Health    []port.HealthChecker `group:"health"`
	Limiter   port.RateLimiter
	Auth      *authn.Authenticator
	Config    *config.Config
}

func NewHandler(p Params) *Handler {
	return &Handler{wishlists: p.Wishlists, health: p.Health, limiter: p.Limiter, auth: p.Auth, writes: p.Config.RateLimit.WritesPerMinute}
}

func (h *Handler) Livez(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	status, code := "ok", http.StatusOK
	components := map[string]string{}
	for _, c := range h.health {
		if err := c.Ping(ctx); err != nil {
			slog.Warn("health check failed", "component", c.Name(), "err", err)
			components[c.Name()] = "unavailable"
			if c.Critical() {
				status, code = "unavailable", http.StatusServiceUnavailable
			} else if status == "ok" {
				status = "degraded"
			}
			continue
		}
		components[c.Name()] = "ok"
	}
	writeJSON(w, code, map[string]any{"status": status, "components": components})
}

func actorFrom(r *http.Request) (model.Actor, bool) {
	p, ok := authn.FromContext(r.Context())
	if !ok {
		return model.Actor{}, false
	}
	return model.Actor{UserID: p.UserID, Admin: p.IsAdmin()}, true
}

func requireActor(w http.ResponseWriter, r *http.Request) (model.Actor, bool) {
	a, ok := actorFrom(r)
	if !ok {
		authn.WriteError(w, authn.ErrUnauthenticated)
	}
	return a, ok
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, &model.ValidationError{Field: name, Reason: "must be a positive integer"})
		return 0, false
	}
	return id, true
}

func queryInt(r *http.Request, name string) (int, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, &model.ValidationError{Field: name, Reason: "must be an integer"}
	}
	return n, nil
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeData(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, map[string]any{"data": v})
}

func writeError(w http.ResponseWriter, err error) {
	var ve *model.ValidationError
	switch {
	case errors.As(err, &ve):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ve.Error()})
	case errors.Is(err, model.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, model.ErrForbidden):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, model.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "conflict"})
	case errors.Is(err, model.ErrUnauthenticated):
		authn.WriteError(w, authn.ErrUnauthenticated)
	case errors.Is(err, model.ErrUnavailable):
		slog.Warn("dependency unavailable", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "dependency unavailable"})
	case errors.Is(err, context.Canceled):
		writeJSON(w, 499, map[string]string{"error": "request canceled"})
	default:
		slog.Error("request failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
