package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/basket-service/internal/adapter/secondary/userservice"
	"github.com/gin-gonic/gin"
)

type fakeAuthn struct{ err error }

func (f fakeAuthn) Authenticate(_ context.Context, token string) (userservice.Identity, error) {
	if f.err != nil {
		return userservice.Identity{}, f.err
	}
	if token != "good" {
		return userservice.Identity{}, userservice.ErrUnauthenticated
	}
	return userservice.Identity{UserID: "42", Role: "admin"}, nil
}

func newAuthTestRouter(authn Authenticator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", RequireAuth(authn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userId": UserID(c)})
	})
	return r
}

func doRequest(r *gin.Engine, header string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequireAuth(t *testing.T) {
	cases := []struct {
		name   string
		authn  Authenticator
		header string
		want   int
	}{
		{"valid token", fakeAuthn{}, "Bearer good", http.StatusOK},
		{"missing header", fakeAuthn{}, "", http.StatusUnauthorized},
		{"not bearer", fakeAuthn{}, "good", http.StatusUnauthorized},
		{"rejected by user-service", fakeAuthn{}, "Bearer bad", http.StatusUnauthorized},
		{"user-service down", fakeAuthn{err: userservice.ErrUpstream}, "Bearer good", http.StatusServiceUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRequest(newAuthTestRouter(tc.authn), tc.header)
			if w.Code != tc.want {
				t.Errorf("status = %d, want %d (body %s)", w.Code, tc.want, w.Body.String())
			}
		})
	}
}
