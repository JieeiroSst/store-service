package router

import (
	"net/http"
	"testing"
)

func TestAdminOnlyRoutes(t *testing.T) {
	fakeUserService(t)
	var h http.Handler = NewRouter(nil, nil, nil, nil)
	for _, rt := range []struct{ method, path string }{
		{"GET", "/v1/partner/"},
		{"DELETE", "/v1/partner/"},
		{"GET", "/debug/pprof/"},
	} {
		if code, _ := doAuth(h, authCall{method: rt.method, path: rt.path}); code != http.StatusUnauthorized {
			t.Errorf("anonymous %s %s: got %d", rt.method, rt.path, code)
		}
		if code, _ := doAuth(h, authCall{method: rt.method, path: rt.path, bearer: testToken("7", "user")}); code != http.StatusForbidden {
			t.Errorf("user %s %s: got %d", rt.method, rt.path, code)
		}
	}
}
