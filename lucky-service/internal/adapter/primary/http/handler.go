package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"github.com/JIeeiroSst/lucky-service/pkg/authn"
	"go.uber.org/fx"
)

const maxBodyBytes = 1 << 20

type Handler struct {
	campaigns port.CampaignUsecase
	games     port.GameUsecase
	checkins  port.CheckinUsecase
	fraud     port.FraudUsecase
	health    []port.HealthChecker
	auth      *authn.Authenticator
}

type Params struct {
	fx.In

	Campaigns port.CampaignUsecase
	Games     port.GameUsecase
	Checkins  port.CheckinUsecase
	Fraud     port.FraudUsecase
	Health    []port.HealthChecker `group:"health"`
	Auth      *authn.Authenticator
}

func NewHandler(p Params) *Handler {
	return &Handler{campaigns: p.Campaigns, games: p.Games, checkins: p.Checkins, fraud: p.Fraud, health: p.Health, auth: p.Auth}
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
	return model.Actor{UserID: p.UserID, Admin: p.IsAdmin(), Service: p.Service}, true
}

func requireActor(w http.ResponseWriter, r *http.Request) (model.Actor, bool) {
	a, ok := actorFrom(r)
	if !ok {
		authn.WriteError(w, authn.ErrUnauthenticated)
	}
	return a, ok
}

func clientInfo(r *http.Request) model.ClientInfo {
	ip := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if ip == "" {
		ip, _, _ = strings.Cut(r.Header.Get("X-Forwarded-For"), ",")
		ip = strings.TrimSpace(ip)
	}
	if ip == "" {
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		}
	}
	return model.ClientInfo{IP: ip, DeviceID: r.Header.Get("X-Device-Id"), UserAgent: r.UserAgent()}
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, &model.ValidationError{Field: name, Reason: "must be a positive integer"})
		return 0, false
	}
	return id, true
}

func paging(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, err := queryInt(r, "limit")
	if err != nil {
		writeError(w, err)
		return 0, 0, false
	}
	offset, err := queryInt(r, "offset")
	if err != nil {
		writeError(w, err)
		return 0, 0, false
	}
	return limit, offset, true
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
		writeJSON(w, http.StatusBadRequest, errorBody("invalid_json", "invalid JSON body: "+err.Error()))
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

func errorBody(code, msg string) map[string]string {
	return map[string]string{"code": code, "error": msg}
}

func writeError(w http.ResponseWriter, err error) {
	var ve *model.ValidationError
	switch {
	case errors.As(err, &ve):
		writeJSON(w, http.StatusBadRequest, errorBody("invalid_argument", ve.Error()))
	case errors.Is(err, model.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, errorBody("invalid_argument", err.Error()))
	case errors.Is(err, model.ErrNoActiveCampaign):
		writeJSON(w, http.StatusNotFound, errorBody("no_active_campaign", err.Error()))
	case errors.Is(err, model.ErrNoSpinsLeft):
		writeJSON(w, http.StatusTooManyRequests, errorBody("no_spins_left", err.Error()))
	case errors.Is(err, model.ErrBlocked):
		writeJSON(w, http.StatusForbidden, errorBody("account_restricted", err.Error()))
	case errors.Is(err, model.ErrDeviceLimit):
		writeJSON(w, http.StatusForbidden, errorBody("device_limit", err.Error()))
	case errors.Is(err, model.ErrDeviceRequired):
		writeJSON(w, http.StatusBadRequest, errorBody("device_required", err.Error()))
	case errors.Is(err, model.ErrAlreadyCheckedIn):
		writeJSON(w, http.StatusConflict, errorBody("already_checked_in", err.Error()))
	case errors.Is(err, model.ErrRateLimited):
		w.Header().Set("Retry-After", "60")
		writeJSON(w, http.StatusTooManyRequests, errorBody("rate_limited", err.Error()))
	case errors.Is(err, model.ErrCreditLimit):
		writeJSON(w, http.StatusTooManyRequests, errorBody("credit_limit_reached", err.Error()))
	case errors.Is(err, model.ErrCampaignOverlap):
		writeJSON(w, http.StatusConflict, errorBody("campaign_overlap", err.Error()))
	case errors.Is(err, model.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody("not_found", "not found"))
	case errors.Is(err, model.ErrForbidden):
		writeJSON(w, http.StatusForbidden, errorBody("forbidden", err.Error()))
	case errors.Is(err, model.ErrConflict):
		writeJSON(w, http.StatusConflict, errorBody("conflict", err.Error()))
	case errors.Is(err, model.ErrUnauthenticated):
		authn.WriteError(w, authn.ErrUnauthenticated)
	case errors.Is(err, model.ErrUnavailable):
		slog.Warn("dependency unavailable", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, errorBody("unavailable", "dependency unavailable"))
	case errors.Is(err, context.Canceled):
		writeJSON(w, 499, errorBody("canceled", "request canceled"))
	default:
		slog.Error("request failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal", "internal server error"))
	}
}
