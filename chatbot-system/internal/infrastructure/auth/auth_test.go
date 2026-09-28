package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"chatbot-system/internal/infrastructure/userservice"
)

type fakeUsers struct{}

func (fakeUsers) Authenticate(_ context.Context, token string) (userservice.Identity, error) {
	switch token {
	case "good":
		return userservice.Identity{UserID: "42"}, nil
	case "down":
		return userservice.Identity{}, userservice.ErrUpstream
	}
	return userservice.Identity{}, userservice.ErrUnauthenticated
}

func TestMiddleware(t *testing.T) {
	var seen int64
	h := New(fakeUsers{}).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = UserID(r.Context())
	}))

	cases := []struct {
		name, header, query string
		want                int
	}{
		{"no token", "", "", http.StatusUnauthorized},
		{"bad token", "Bearer bad", "", http.StatusUnauthorized},
		{"user-service down", "Bearer down", "", http.StatusServiceUnavailable},
		{"header token", "Bearer good", "", http.StatusOK},
		{"websocket query token", "", "?token=good", http.StatusOK},
		{"client-supplied user_id is ignored", "", "?user_id=1", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		seen = 0
		req := httptest.NewRequest(http.MethodGet, "/api/conversations"+tc.query, nil)
		if tc.header != "" {
			req.Header.Set("Authorization", tc.header)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s: status %d, want %d", tc.name, w.Code, tc.want)
		}
		if tc.want == http.StatusOK && seen != 42 {
			t.Errorf("%s: user id %d, want 42", tc.name, seen)
		}
	}
}
