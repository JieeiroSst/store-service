// Package auth identifies callers through user-service, which owns every
// account and issues every bearer token.
package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"chatbot-system/internal/infrastructure/userservice"
)

type ctxKey struct{}

// Identities resolves a bearer token to a user-service identity.
type Identities interface {
	Authenticate(ctx context.Context, token string) (userservice.Identity, error)
}

type Authenticator struct {
	users Identities
}

func New(users Identities) *Authenticator { return &Authenticator{users: users} }

// Authenticate reads the token from the Authorization header, or the
// "token" query parameter for WebSocket upgrades, and returns the caller's
// user-service id. On failure it has already written the HTTP error.
func (a *Authenticator) Authenticate(w http.ResponseWriter, r *http.Request) (int64, bool) {
	token := r.URL.Query().Get("token")
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		token = strings.TrimPrefix(h, "Bearer ")
	}
	id, err := a.users.Authenticate(r.Context(), token)
	if errors.Is(err, userservice.ErrUpstream) {
		http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
		return 0, false
	}
	userID, convErr := strconv.ParseInt(id.UserID, 10, 64)
	if err != nil || convErr != nil {
		http.Error(w, "invalid or missing bearer token", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

// Middleware rejects unauthenticated requests and stores the caller's id.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := a.Authenticate(w, r)
		if !ok {
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, userID)))
	})
}

// UserID is the authenticated caller set by Middleware.
func UserID(ctx context.Context) int64 {
	id, _ := ctx.Value(ctxKey{}).(int64)
	return id
}
