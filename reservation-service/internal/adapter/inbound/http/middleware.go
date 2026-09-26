package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
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

func (h *Handler) authenticate(r *http.Request) (*http.Request, error) {
	tok := bearer(r)
	if tok == "" {
		return r, errUnauthorized
	}
	p, err := h.auth.Authenticate(r.Context(), tok)
	if err != nil {
		return r, err
	}
	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, p)), nil
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := h.authenticate(r)
		if err != nil {
			h.fail(w, err)
			return
		}
		next(w, r)
	})
}

func limitInFlight(max int, next http.Handler) http.Handler {
	if max <= 0 {
		return next
	}
	slots := make(chan struct{}, max)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			next.ServeHTTP(w, r)
		default:
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"the service is overloaded, try again in a moment"}`))
		}
	})
}

func logging(log *slog.Logger, next http.Handler) http.Handler {
	var seen atomic.Uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if v := recover(); v != nil {
				log.Error("panic", "panic", v, "path", r.URL.Path)
				http.Error(rec, `{"error":"internal error"}`, http.StatusInternalServerError)
				rec.status = http.StatusInternalServerError
			}
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
				return
			}
			dur := time.Since(start)
			attrs := []any{"method", r.Method, "path", r.URL.Path, "status", rec.status, "dur", dur}
			switch {
			case rec.status >= 500:
				log.Error("http", attrs...)
			case dur > 500*time.Millisecond:
				log.Warn("http slow", attrs...)
			case rec.status >= 400 || r.Method == http.MethodGet:
				if seen.Add(1)%100 == 1 {
					log.Info("http (1 in 100 shown)", attrs...)
				}
			default:
				log.Info("http", attrs...)
			}
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
