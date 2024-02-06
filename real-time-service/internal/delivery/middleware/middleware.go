package middleware

import (
	"net/http"

	"github.com/JIeeiroSst/real-time-service/config"
	"github.com/JIeeiroSst/real-time-service/pkg/logger"
)

type MiddlewareDelivery struct {
	config *config.Config
}

func NewMiddlewareDelivery(config *config.Config) *MiddlewareDelivery {
	return &MiddlewareDelivery{
		config: config,
	}
}

func (m *MiddlewareDelivery) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var apiKey string
		if apiKey = req.Header.Get("X-Api-Key"); apiKey != m.config.Serect.Key {
			logger.Logger().Sugar().Infof("bad auth api key: %s", apiKey)
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
