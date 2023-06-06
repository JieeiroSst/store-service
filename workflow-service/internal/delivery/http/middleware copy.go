package http

import (
	"log"
	"net/http"
	"time"

	"github.com/JIeeiroSst/workflow-service/utils"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func middlewareTwo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("Executing middlewareTwo")
		if r.URL.Path == "/foo" {
			return
		}

		next.ServeHTTP(w, r)
		log.Print("Executing middlewareTwo again")
	})
}

func (h *Http) corsMiddleware(router chi.Router) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(h.AccessApi)
}

const requestIDHeader = "Authorization"

func (h *Http) AccessApi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorzation := w.Header().Get(requestIDHeader)
		if ok := utils.DecodeBase(authorzation, h.Config.Secret.AuthorizeKey); !ok {
			http.Error(w, http.StatusText(403), 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}
