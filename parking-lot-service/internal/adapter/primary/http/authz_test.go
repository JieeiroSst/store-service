package http

import "testing"

func TestParkingAuthorization(t *testing.T) {
	fakeUserService(t)
	router := NewRouter(&Handler{})
	user := testToken("7", "user")
	for _, c := range []struct {
		name string
		call authCall
		want int
	}{
		{"anonymous check-in", authCall{method: "POST", path: "/api/v1/check-in", body: `{}`}, 401},
		{"user check-in", authCall{method: "POST", path: "/api/v1/check-in", body: `{}`, bearer: user}, 403},
		{"user ticket", authCall{method: "GET", path: "/api/v1/tickets/1", bearer: user}, 403},
		{"user history", authCall{method: "GET", path: "/api/v1/history?plate=x", bearer: user}, 403},
	} {
		if got, body := doAuth(router, c.call); got != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, got, c.want, body)
		}
	}
	if got, _ := doAuth(router, authCall{method: "GET", path: "/api/v1/rates"}); got == 401 || got == 403 {
		t.Errorf("public rates blocked: %d", got)
	}
}
