package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/userservice"
	"github.com/gin-gonic/gin"
)

type fakeAuthn map[string]userservice.Identity

func (f fakeAuthn) Authenticate(_ context.Context, token string) (userservice.Identity, error) {
	if id, ok := f[token]; ok {
		return id, nil
	}
	if token == "down" {
		return userservice.Identity{}, userservice.ErrUpstream
	}
	return userservice.Identity{}, userservice.ErrUnauthenticated
}

func TestAuthMiddlewareRolesAndPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authn := fakeAuthn{
		"alice": {UserID: "1", Role: "user"},
		"root":  {UserID: "2", Role: "super_admin"},
		"aud":   {UserID: "3", Role: "auditor"},
	}
	r := gin.New()
	r.Use(AuthMiddleware(authn))
	r.POST("/keys", RequirePermission("key:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.MustGet("user_id").(int64)})
	})
	r.GET("/audit", RequireRole(models.RoleAuditor), func(c *gin.Context) { c.Status(http.StatusOK) })

	cases := []struct {
		method, path, token string
		want                int
	}{
		{http.MethodPost, "/keys", "", http.StatusUnauthorized},
		{http.MethodPost, "/keys", "bad", http.StatusUnauthorized},
		{http.MethodPost, "/keys", "down", http.StatusServiceUnavailable},
		{http.MethodPost, "/keys", "alice", http.StatusOK},
		{http.MethodPost, "/keys", "aud", http.StatusForbidden},
		{http.MethodPost, "/keys", "root", http.StatusOK},
		{http.MethodGet, "/audit", "alice", http.StatusForbidden},
		{http.MethodGet, "/audit", "aud", http.StatusOK},
		{http.MethodGet, "/audit", "root", http.StatusOK},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s %s as %q = %d, want %d", tc.method, tc.path, tc.token, w.Code, tc.want)
		}
	}
}
