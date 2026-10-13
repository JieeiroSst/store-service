package http

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminOnlyRoutes(t *testing.T) {
	fakeUserService(t)
	var h http.Handler = func() http.Handler { r := gin.New(); SetupRoutes(r, nil, nil); return r }()
	for _, rt := range []struct{ method, path string }{
		{"GET", "/api/v1/circle/wallets"},
		{"POST", "/api/v1/bridge/solana-to-circle"},
	} {
		if code, _ := doAuth(h, authCall{method: rt.method, path: rt.path}); code != http.StatusUnauthorized {
			t.Errorf("anonymous %s %s: got %d", rt.method, rt.path, code)
		}
		if code, _ := doAuth(h, authCall{method: rt.method, path: rt.path, bearer: testToken("7", "user")}); code != http.StatusForbidden {
			t.Errorf("user %s %s: got %d", rt.method, rt.path, code)
		}
	}
}
