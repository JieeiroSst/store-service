package http

import (
	"net/http"

	"github.com/JIeeiroSst/webrtc-service/web"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.Index)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/ice-servers", h.ICEServers)
	mux.HandleFunc("GET /api/rooms/{room_id}", h.Room)
	mux.HandleFunc("GET /ws", h.ServeWS)
	mux.Handle("GET /static/", immutable(http.FileServerFS(web.Static)))
	return mux
}

func immutable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
