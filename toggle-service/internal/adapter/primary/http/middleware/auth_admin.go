package middleware

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"

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
			if errors.Is(err, apperr.ErrUnauthorized) {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			if err != nil {
				http.Error(w, `{"error":"authentication service unavailable"}`, http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), userID, isAdmin)))
		})
	}
}
