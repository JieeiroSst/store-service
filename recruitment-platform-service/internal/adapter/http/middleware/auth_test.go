package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/userservice"
)

type fakeAuthn struct{}

func (fakeAuthn) Authenticate(_ context.Context, token string) (userservice.Identity, error) {
	switch token {
	case "good":
		return userservice.Identity{UserID: "42", Role: "operator"}, nil
	case "down":
		return userservice.Identity{}, userservice.ErrUpstream
	}
	return userservice.Identity{}, userservice.ErrUnauthenticated
}

func TestUserServiceAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", UserServiceAuth(fakeAuthn{}, zap.NewNop()), func(c *gin.Context) {
		if c.MustGet("user_id").(int64) != 42 || c.MustGet("role").(string) != "operator" {
			t.Errorf("unexpected identity in context")
		}
		c.Status(http.StatusOK)
	})

	for header, want := range map[string]int{
		"":            http.StatusUnauthorized,
		"good":        http.StatusUnauthorized, // missing "Bearer "
		"Bearer bad":  http.StatusUnauthorized,
		"Bearer down": http.StatusServiceUnavailable,
		"Bearer good": http.StatusOK,
	} {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != want {
			t.Errorf("Authorization %q: status %d, want %d", header, w.Code, want)
		}
	}
}
