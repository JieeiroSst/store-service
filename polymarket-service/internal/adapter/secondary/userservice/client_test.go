package userservice

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func client(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cfg := &config.Config{UserService: config.UserServiceConfig{BaseURL: srv.URL, Timeout: "2s"}}
	return New(cfg)
}

func TestValidateSendsTheTokenAndReturnsTheUser(t *testing.T) {
	c := client(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/validate" || body["session_token"] != "tok" {
			t.Errorf("unexpected request %s %s %v", r.Method, r.URL.Path, body)
		}
		_, _ = w.Write([]byte(`{"valid":true,"user_id":"17"}`))
	})
	if id, err := c.Validate(ctx(), "tok"); err != nil || id != "17" {
		t.Fatalf("%q %v", id, err)
	}
}

func TestValidateAcceptsCamelCase(t *testing.T) {
	c := client(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"valid":true,"userId":"9"}`)) })
	if id, err := c.Validate(ctx(), "tok"); err != nil || id != "9" {
		t.Fatalf("%q %v", id, err)
	}
}

func TestValidateRefusesInvalidTokens(t *testing.T) {
	for name, h := range map[string]http.HandlerFunc{
		"valid=false": func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"valid":false}`)) },
		"no user id":  func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"valid":true}`)) },
		"401":         func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) },
	} {
		if _, err := client(t, h).Validate(ctx(), "tok"); !errors.Is(err, port.ErrUnauthenticated) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestValidateReportsUpstreamFailuresDistinctly(t *testing.T) {
	c := client(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	if _, err := c.Validate(ctx(), "tok"); !errors.Is(err, port.ErrUpstream) {
		t.Fatalf("a user-service outage must not look like a bad token: %v", err)
	}
	unconfigured := New(&config.Config{})
	if _, err := unconfigured.Validate(ctx(), "tok"); !errors.Is(err, port.ErrUpstream) {
		t.Fatalf("%v", err)
	}
}
