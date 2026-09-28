package http

import (
	"net/http"

	"github.com/JIeeiroSst/manage-service/utils"
)

const apiKeyHeader = "X-Api-Key"

func apiKeyMiddleware(authorizeKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authorizeKey != "" && !utils.DecodeBase(r.Header.Get(apiKeyHeader), authorizeKey) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
