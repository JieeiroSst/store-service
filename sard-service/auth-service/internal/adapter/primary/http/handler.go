package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/auth-service/config"
	"github.com/JIeeiroSst/auth-service/internal/application"
	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/JIeeiroSst/auth-service/internal/port"
)

const (
	maxBodyBytes  = 16 << 10
	healthTimeout = 2 * time.Second
)

type Handler struct {
	auth     *application.AuthenticationService
	health   port.HealthChecker
	merchant tokenSet
	internal tokenSet
	limiter  *limiter
}

func NewHandler(cfg *config.Config, auth *application.AuthenticationService, health port.HealthChecker) *Handler {
	return &Handler{
		auth:     auth,
		health:   health,
		merchant: newTokenSet(cfg.Server.AccessTokens),
		internal: newTokenSet(cfg.Server.InternalTokens),
		limiter:  newLimiter(cfg.Server.VerifyPerMinute, time.Minute),
	}
}

type challengeDTO struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	CardID          string     `json:"card_id"`
	MaskedPAN       string     `json:"masked_pan"`
	Amount          int64      `json:"amount"`
	Currency        string     `json:"currency"`
	Merchant        string     `json:"merchant"`
	MCC             string     `json:"mcc,omitempty"`
	AttemptsLeft    int        `json:"attempts_left"`
	Resends         int        `json:"resends"`
	OTPExpiresAt    time.Time  `json:"otp_expires_at"`
	FailureReason   string     `json:"failure_reason,omitempty"`
	ExpiresAt       time.Time  `json:"expires_at"`
	AuthenticatedAt *time.Time `json:"authenticated_at,omitempty"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	UsedAt          *time.Time `json:"used_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	OTP             string     `json:"otp,omitempty"`
}

func toChallenge(c *domain.Challenge) challengeDTO {
	left := c.MaxAttempts - c.Attempts
	if left < 0 {
		left = 0
	}
	return challengeDTO{
		ID: c.ID, Status: string(c.Status), CardID: c.CardID, MaskedPAN: c.MaskedPAN, Amount: c.Amount,
		Currency: c.Currency, Merchant: c.Merchant, MCC: c.MCC, AttemptsLeft: left, Resends: c.Resends,
		OTPExpiresAt: c.OTPExpiresAt, FailureReason: c.FailureReason,
		ExpiresAt: c.ExpiresAt, AuthenticatedAt: c.AuthenticatedAt, ValidUntil: c.ValidUntil, UsedAt: c.UsedAt,
		CreatedAt: c.CreatedAt,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()
	if err := h.health.Ping(ctx); err != nil {
		log.Printf("health check failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "version": config.Version})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": config.Version})
}

func (h *Handler) Initiate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PAN      string `json:"pan"`
		Expiry   string `json:"expiry"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
		Merchant string `json:"merchant"`
		MCC      string `json:"mcc"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	out, err := h.auth.Initiate(r.Context(), domain.InitiateRequest{
		PAN: req.PAN, Expiry: req.Expiry, Amount: req.Amount, Currency: req.Currency, Merchant: req.Merchant, MCC: req.MCC,
	})
	if err != nil {
		fail(w, err)
		return
	}
	dto := toChallenge(out.Challenge)
	dto.OTP = out.OTP
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, dto)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	c, err := h.auth.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toChallenge(c))
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	list, err := h.auth.ListPending(r.Context(), sessionToken(r))
	if err != nil {
		fail(w, err)
		return
	}
	out := make([]challengeDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toChallenge(c))
	}
	writeJSON(w, http.StatusOK, map[string]any{"payment_authentications": out, "count": len(out)})
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OTP string `json:"otp"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.auth.Verify(r.Context(), r.PathValue("id"), sessionToken(r), req.OTP)
	if err != nil {
		if c != nil && (domain.IsOTPMismatch(err) || errors.Is(err, domain.ErrExpired)) {
			status, msg := errorStatus(err)
			writeJSON(w, status, map[string]any{"error": msg, "payment_authentication": toChallenge(c)})
			return
		}
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toChallenge(c))
}

func (h *Handler) Resend(w http.ResponseWriter, r *http.Request) {
	out, err := h.auth.Resend(r.Context(), r.PathValue("id"), sessionToken(r))
	if err != nil {
		fail(w, err)
		return
	}
	dto := toChallenge(out.Challenge)
	dto.OTP = out.OTP
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, dto)
}

func (h *Handler) Decline(w http.ResponseWriter, r *http.Request) {
	c, err := h.auth.Decline(r.Context(), r.PathValue("id"), sessionToken(r))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toChallenge(c))
}

func (h *Handler) Consume(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CardID   string `json:"card_id"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := decodeBody(r, &req); err != nil {
		fail(w, err)
		return
	}
	c, err := h.auth.Consume(r.Context(), r.PathValue("id"), application.ConsumeCommand{
		CardID: req.CardID, Amount: req.Amount, Currency: req.Currency,
	})
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toChallenge(c))
}

func sessionToken(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return strings.TrimSpace(r.Header.Get("X-Session-Token"))
}

func fail(w http.ResponseWriter, err error) {
	status, msg := errorStatus(err)
	if status >= 500 {
		log.Printf("request failed: %v", err)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func errorStatus(err error) (int, string) {
	switch {
	case domain.IsInvalid(err):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "missing or invalid token"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "payment authentication not found"
	case domain.IsOTPMismatch(err):
		return http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, domain.ErrExpired), errors.Is(err, domain.ErrOTPExpired):
		return http.StatusGone, err.Error()
	case errors.Is(err, domain.ErrOTPLimit):
		return http.StatusTooManyRequests, err.Error()
	case domain.IsConflict(err):
		return http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrUnavailable):
		return http.StatusServiceUnavailable, "a dependency is unavailable, try again later"
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "request timed out"
	}
	return http.StatusInternalServerError, "internal error"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeBody(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return domain.Invalid("invalid JSON body: %v", err)
	}
	return nil
}
