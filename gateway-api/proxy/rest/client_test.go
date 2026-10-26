package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/gateway-api/proxy/breaker"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestForwardsUserAccessToken(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()
	user := auth.WithUser(context.Background(), auth.User{ID: "7", Token: "session-token-from-login"})

	var out any
	if err := New("wishlists-service", srv.URL, time.Second).Get(user, "/api/v1/wishlists", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "Bearer session-token-from-login" || got.Get("X-User-Id") != "7" {
		t.Fatalf("user access token must be forwarded: %v", got)
	}

	if err := New("wishlists-service", srv.URL, time.Second).Get(context.Background(), "/api/v1/wishlist-stats/popular-products", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "" || got.Get("X-User-Id") != "" {
		t.Fatalf("anonymous calls must not carry credentials: %v", got)
	}

	if err := New("search-service", srv.URL, time.Second, Options{DropUserToken: true}).Get(user, "/", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "" {
		t.Fatalf("forward_user_token=false must drop the token: %v", got)
	}
}

func TestForwardsClientSignals(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()
	ctx := auth.WithClient(context.Background(), auth.Client{IP: "203.0.113.5", UserAgent: "LuckyApp/2", DeviceID: "device-9"})
	var out any
	if err := New("lucky-service", srv.URL, time.Second).Post(ctx, "/api/v1/lucky/checkin/me", map[string]any{}, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("X-Real-IP") != "203.0.113.5" || got.Get("User-Agent") != "LuckyApp/2" || got.Get("X-Device-Id") != "device-9" {
		t.Fatalf("client ip, user agent and device id must reach the service: %v", got)
	}
	if err := New("lucky-service", srv.URL, time.Second).Get(context.Background(), "/", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("X-Device-Id") != "" {
		t.Fatalf("no device id without a client: %v", got)
	}
}

func TestUpstreamErrorReasonAndDetails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"AUTH_MFA_INVALID_CODE","message":"authentication code is invalid or already used","status":401,"details":{"k":"v"}}`))
	}))
	defer srv.Close()

	var out any
	err := New("user-service", srv.URL, time.Second).Post(context.Background(), "/api/v1/mfa/totp/verify", map[string]string{}, &out)
	ge, ok := err.(*gqlerror.Error)
	if !ok {
		t.Fatalf("err = %T %v", err, err)
	}
	if ge.Message != "authentication code is invalid or already used" || ge.Extensions["code"] != auth.CodeUnauthenticated || ge.Extensions["reason"] != "AUTH_MFA_INVALID_CODE" {
		t.Fatalf("extensions = %v message = %q", ge.Extensions, ge.Message)
	}
	if d, _ := ge.Extensions["details"].(map[string]string); d["k"] != "v" {
		t.Fatalf("details = %v", ge.Extensions["details"])
	}
}

func TestCircuitBreakerShortCircuitsFailingService(t *testing.T) {
	hits := 0
	status := http.StatusInternalServerError
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(status)
	}))
	defer srv.Close()
	c := New("flaky-service", srv.URL, time.Second, Options{Breaker: breaker.New("flaky-service", breaker.Settings{FailureThreshold: 2, OpenTimeout: time.Hour})})

	for i := 0; i < 2; i++ {
		_ = c.Get(context.Background(), "/", nil, nil)
	}
	status = http.StatusNotFound
	err := c.Get(context.Background(), "/", nil, nil)
	if hits != 2 {
		t.Fatalf("open breaker must not reach the service, hits=%d", hits)
	}
	ge, ok := err.(*gqlerror.Error)
	if !ok || ge.Extensions["code"] != auth.CodeServiceUnavailable || ge.Extensions["status"] != http.StatusServiceUnavailable {
		t.Fatalf("want SERVICE_UNAVAILABLE 503, got %v", err)
	}
}

func TestCircuitBreakerIgnoresClientErrors(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()
	c := New("strict-service", srv.URL, time.Second, Options{Breaker: breaker.New("strict-service", breaker.Settings{FailureThreshold: 1})})
	for i := 0; i < 3; i++ {
		_ = c.Get(context.Background(), "/", nil, nil)
	}
	if hits != 3 {
		t.Fatalf("4xx must not trip the breaker, hits=%d", hits)
	}
}
