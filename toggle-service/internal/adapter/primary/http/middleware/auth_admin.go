package middleware

import (
	"net/http"

	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

func RequireAdminAuth(authService port.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractBearerToken(r)
			if tokenString == "" {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}
			userID, isAdmin, err := authService.VerifyToken(r.Context(), tokenString)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), userID, isAdmin)))
		})
	}
}
