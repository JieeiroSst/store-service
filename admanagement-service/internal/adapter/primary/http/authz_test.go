package http

import (
	"testing"

	"github.com/gorilla/mux"
)

func TestAdmanagementAuthorization(t *testing.T) {
	fakeUserService(t)
	r := mux.NewRouter()
	RegisterRoutes(r, &Handler{})
	user := testToken("7", "user")
	for _, c := range []struct {
		name string
		call authCall
		want int
	}{
		{"anonymous campaigns", authCall{method: "GET", path: "/campaigns"}, 401},
		{"user campaigns", authCall{method: "GET", path: "/campaigns", bearer: user}, 403},
		{"user creates ad", authCall{method: "POST", path: "/ads", body: `{}`, bearer: user}, 403},
		{"user impressions", authCall{method: "GET", path: "/impressions", bearer: user}, 403},
	} {
		if got, body := doAuth(r, c.call); got != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, got, c.want, body)
		}
	}
}
