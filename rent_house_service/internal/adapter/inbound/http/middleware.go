package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
)

type ctxKey struct{}

func principalFrom(ctx context.Context) (inbound.Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(inbound.Principal)
	return p, ok
}

func mustPrincipal(r *http.Request) inbound.Principal {
	p, _ := principalFrom(r.Context())
	return p
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func (h *Handler) optionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tok := bearer(r); tok != "" {
			if p, err := h.auth.Authenticate(r.Context(), tok); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, p))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := bearer(r)
		if tok == "" {
			h.fail(w, errUnauthorized)
			return
		}
		p, err := h.auth.Authenticate(r.Context(), tok)
		if err != nil {
			h.fail(w, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, p)))
	})
}

func logging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if v := recover(); v != nil {
				log.Error("panic", "panic", v, "path", r.URL.Path)
				http.Error(rec, `{"error":"internal error"}`, http.StatusInternalServerError)
			}
			log.Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status, "dur", time.Since(start))
		}()
		next.ServeHTTP(rec, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(c int) {
	s.status = c
	s.ResponseWriter.WriteHeader(c)
}
