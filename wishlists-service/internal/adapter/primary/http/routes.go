package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/wishlists-service/pkg/authn"
	"github.com/JIeeiroSst/wishlists-service/pkg/credential"
)

func (h *Handler) Routes() http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("POST /api/v1/wishlists", h.CreateWishlist)
	api.HandleFunc("GET /api/v1/wishlists", h.ListWishlists)
	api.HandleFunc("GET /api/v1/wishlists/{id}", h.GetWishlist)
	api.HandleFunc("PUT /api/v1/wishlists/{id}", h.UpdateWishlist)
	api.HandleFunc("PATCH /api/v1/wishlists/{id}", h.UpdateWishlist)
	api.HandleFunc("DELETE /api/v1/wishlists/{id}", h.DeleteWishlist)
	api.HandleFunc("POST /api/v1/wishlists/{id}/lines", h.AddLine)
	api.HandleFunc("PUT /api/v1/wishlists/{id}/lines/{lineId}", h.UpdateLine)
	api.HandleFunc("PATCH /api/v1/wishlists/{id}/lines/{lineId}", h.UpdateLine)
	api.HandleFunc("DELETE /api/v1/wishlists/{id}/lines", h.ClearLines)
	api.HandleFunc("DELETE /api/v1/wishlists/{id}/lines/{lineId}", h.RemoveLine)
	api.HandleFunc("POST /api/v1/wishlists/{id}/lines/{lineId}/move", h.MoveLine)
	api.HandleFunc("GET /api/v1/shared-wishlists/{key}", h.GetSharedWishlist)
	api.HandleFunc("POST /api/v1/shared-wishlists/{key}/copy", h.CopySharedWishlist)
	api.HandleFunc("GET /api/v1/users/{userId}/wishlists", h.ListUserWishlists)
	api.HandleFunc("GET /api/v1/me/wishlists/default", h.DefaultWishlist)
	api.HandleFunc("POST /api/v1/me/wishlist-items", h.AddToDefault)
	api.HandleFunc("GET /api/v1/me/products/{productId}/wishlists", h.FindProduct)
	api.HandleFunc("GET /api/v1/wishlist-stats/popular-products", h.PopularProducts)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", h.Livez)
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("/api/", h.auth.OptionalMiddleware(forwardCredentials(h.rateLimit(api))))
	return recoverer(accessLog(mux))
}

func (h *Handler) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		p, ok := authn.FromContext(r.Context())
		if !ok || p.Service || p.UserID == "" {
			next.ServeHTTP(w, r)
			return
		}
		allowed, err := h.limiter.Allow(r.Context(), "user:"+p.UserID, h.writes, time.Minute)
		if err != nil {
			slog.Warn("rate limiter unavailable, allowing request", "err", err)
		}
		if !allowed {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func forwardCredentials(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := authn.FromContext(r.Context()); ok {
			r = r.WithContext(credential.With(r.Context(), credential.FromRequest(r)))
		}
		next.ServeHTTP(w, r)
	})
}

func redactPath(p string) string {
	const shared = "/api/v1/shared-wishlists/"
	rest, ok := strings.CutPrefix(p, shared)
	if !ok || rest == "" {
		return p
	}
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		return shared + "{key}" + rest[i:]
	}
	return shared + "{key}"
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if r.URL.Path == "/livez" || r.URL.Path == "/health" {
			return
		}
		slog.Info("http request", "method", r.Method, "path", redactPath(r.URL.Path), "status", rec.status, "duration", time.Since(start))
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				slog.Error("panic while serving request", "panic", v, "path", redactPath(r.URL.Path))
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
