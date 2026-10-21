package ollamaapi

import "net/http"

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/generate", h.Generate)
	mux.HandleFunc("POST /api/chat", h.Chat)
	mux.HandleFunc("GET /api/tags", h.Tags)
	mux.HandleFunc("POST /api/show", h.Show)
	return mux
}
