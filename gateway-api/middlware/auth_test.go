package middlware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/gateway-api/proxy/user_proxy"
	"github.com/gin-gonic/gin"
)

type fakeUsers struct {
	user_proxy.UserProxy
}

func (fakeUsers) ValidateSession(_ context.Context, token string) (user_proxy.Session, error) {
	switch token {
	case "good":
		return user_proxy.Session{Valid: true, UserID: "7"}, nil
	case "down":
		return user_proxy.Session{}, errors.New("unavailable")
	}
	return user_proxy.Session{Valid: false}, nil
}

func TestAuthenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var seen auth.User
	var authed bool
	r := gin.New()
	r.Use(Authenticate(fakeUsers{}))
	r.GET("/", func(c *gin.Context) {
		seen, authed = auth.UserFromContext(c.Request.Context())
		c.Status(http.StatusOK)
	})
	call := func(header string) int {
		authed, seen = false, auth.User{}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := call("Bearer good"); code != 200 || !authed || seen.ID != "7" || seen.Token != "good" {
		t.Fatalf("valid access token: %d %+v", code, seen)
	}
	if code := call(""); code != 200 || authed {
		t.Fatalf("anonymous request: %d %v", code, authed)
	}
	if code := call("Bearer expired"); code != http.StatusUnauthorized {
		t.Fatalf("invalid token: %d", code)
	}
	if code := call("Basic abc"); code != http.StatusUnauthorized {
		t.Fatalf("wrong scheme: %d", code)
	}
	if code := call("Bearer down"); code != http.StatusServiceUnavailable {
		t.Fatalf("user-service down: %d", code)
	}
}
