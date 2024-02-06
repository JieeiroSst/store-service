package delivery

import (
	"log"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	TestApiKey := "test_api_key"
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var apiKey string
		if apiKey = req.Header.Get("X-Api-Key"); apiKey != TestApiKey {
			log.Printf("bad auth api key: %s", apiKey)
			rw.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(rw, req)
	})
}

func Middleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}