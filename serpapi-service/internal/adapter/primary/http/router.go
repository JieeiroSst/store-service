package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/engines", h.Engines)
	api.HandleFunc("GET /api/v1/engines/{engine}", h.Engine)

	api.HandleFunc("GET /api/v1/search", h.SearchQuery)
	api.HandleFunc("GET /api/v1/search/{engine}", h.SearchQuery)
	api.HandleFunc("POST /api/v1/search", h.SearchBody)
	api.HandleFunc("POST /api/v1/search/batch", h.SearchBatch)
	api.HandleFunc("GET /api/v1/searches/{id}", h.Archive)

	api.HandleFunc("GET /api/v1/account", h.Account)
	api.HandleFunc("GET /api/v1/locations", h.Locations)
	api.HandleFunc("POST /api/v1/images", h.UploadImage)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("/api/", h.authorize(api))
	return mux
}
