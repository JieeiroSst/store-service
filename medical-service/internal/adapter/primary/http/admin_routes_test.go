package http

import (
	"net/http"
	"testing"
)

func TestAdminOnlyRoutes(t *testing.T) {
	fakeUserService(t)
	var h http.Handler = NewRouter(&Handler{})
	for _, rt := range []struct{ method, path string }{
		{"GET", "/api/v1/medicines"},
		{"POST", "/api/v1/dispenses"},
	} {
		if code, _ := doAuth(h, authCall{method: rt.method, path: rt.path}); code != http.StatusUnauthorized {
			t.Errorf("anonymous %s %s: got %d", rt.method, rt.path, code)
		}
		if code, _ := doAuth(h, authCall{method: rt.method, path: rt.path, bearer: testToken("7", "user")}); code != http.StatusForbidden {
			t.Errorf("user %s %s: got %d", rt.method, rt.path, code)
		}
	}
}
