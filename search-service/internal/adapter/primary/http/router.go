package http

import (
	"net/http"
	"time"
)

const apiTimeout = 30 * time.Second

func NewRouter(h *Handler) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/search", h.Search)
	api.HandleFunc("GET /api/v1/autocomplete", h.Autocomplete)
	api.HandleFunc("GET /api/v1/documents", h.List)
	api.HandleFunc("GET /api/v1/documents/{index}/{id}", h.Get)
	api.HandleFunc("GET /api/v1/documents/{index}/{id}/similar", h.Similar)
	api.HandleFunc("GET /api/v1/indices", h.Indices)
	api.HandleFunc("GET /api/v1/indices/{index}/fields", h.Fields)

	admin := http.NewServeMux()
	admin.HandleFunc("GET /api/v1/admin/schema", h.SchemaStatus)
	admin.HandleFunc("PUT /api/v1/admin/schema", h.ApplySchema)
	admin.HandleFunc("POST /api/v1/admin/reindex", h.Reindex)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /livez", h.Live)
	mux.Handle("/api/v1/admin/", h.authorize(admin))
	mux.Handle("/api/", h.authorize(http.TimeoutHandler(api, apiTimeout, `{"error":"request timed out"}`)))
	return mux
}
