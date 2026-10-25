package http

import (
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"github.com/JIeeiroSst/gateway-service/pkg/authn"
	"github.com/gin-gonic/gin"
)

type fakeGateway struct{ h nethttp.Handler }

func (f fakeGateway) ConsumerHandler() nethttp.Handler   { return f.h }
func (f fakeGateway) AccountingHandler() nethttp.Handler { return f.h }
func (f fakeGateway) DeliveryHandler() nethttp.Handler   { return f.h }

func TestGatewayRequiresUserServiceAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	users := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var in struct {
			SessionToken string `json:"session_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.SessionToken == "access-7" {
			_, _ = w.Write([]byte(`{"valid":true,"user_id":"7"}`))
			return
		}
		_, _ = w.Write([]byte(`{"valid":false}`))
	}))
	defer users.Close()

	var seen nethttp.Header
	var path string
	backend := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		seen, path = r.Header.Clone(), r.URL.Path
		w.WriteHeader(nethttp.StatusOK)
	}))
	defer backend.Close()
	target, _ := url.Parse(backend.URL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	newRouter := func(tokens []string) string {
		srv := httptest.NewServer(NewRouter(NewHandler(fakeGateway{proxy}, authn.New(users.URL, tokens, time.Second))))
		t.Cleanup(srv.Close)
		return srv.URL
	}
	get := func(base, p string, header ...string) int {
		seen, path = nil, ""
		req, _ := nethttp.NewRequest(nethttp.MethodGet, base+p, nil)
		for i := 0; i+1 < len(header); i += 2 {
			req.Header.Set(header[i], header[i+1])
		}
		resp, err := nethttp.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	call := func(base string, header ...string) int {
		return get(base, "/consumer/api/v1/consumer", header...)
	}

	enforcing := newRouter([]string{"svc"})
	if code := call(enforcing, "Authorization", "Bearer access-7", "X-User-Id", "1"); code != 200 {
		t.Fatalf("valid access token: %d", code)
	}
	if seen.Get("Authorization") != "Bearer access-7" || seen.Get("X-User-Id") != "7" || path != "/api/v1/consumer" {
		t.Fatalf("downstream must get the user's token and the verified user id: %v %s", seen, path)
	}
	if code := call(enforcing); code != nethttp.StatusUnauthorized || seen != nil {
		t.Fatalf("anonymous request must be rejected: %d", code)
	}
	if code := call(enforcing, "Authorization", "Bearer forged"); code != nethttp.StatusUnauthorized || seen != nil {
		t.Fatalf("invalid token must be rejected: %d", code)
	}
	if code := call(enforcing, "X-Internal-Token", "svc"); code != 200 {
		t.Fatalf("service token: %d", code)
	}
	if code := call(newRouter(nil), "X-User-Id", "1"); code != nethttp.StatusUnauthorized || seen != nil {
		t.Fatalf("without internal tokens anonymous requests must still be rejected: %d %v", code, seen)
	}
	if code := get(enforcing, "/health"); code != 200 {
		t.Fatalf("health must stay open: %d", code)
	}
}
