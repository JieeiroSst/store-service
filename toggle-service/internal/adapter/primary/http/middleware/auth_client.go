package middleware

import (
	"net/http"

	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

func RequireClientToken(tokenService port.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractBearerToken(r)
			if tokenString == "" {
				http.Error(w, `{"error":"missing API token"}`, http.StatusUnauthorized)
				return
			}
			tok, err := tokenService.Resolve(r.Context(), tokenString)
			if err != nil || tok == nil {
				http.Error(w, `{"error":"invalid API token"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithAPIToken(r.Context(), tok)))
		})
	}
}
