package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouterConfig struct {
	AuthorizeKey string
	Timeout      time.Duration
}

func NewRouter(h *Handler, cfg RouterConfig) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(cfg.Timeout))

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Group(func(router chi.Router) {
		router.Use(apiKeyMiddleware(cfg.AuthorizeKey))
		h.authRoutes(router)
		router.Route("/api/v1", h.keycloakRoutes)
	})
	return router
}
