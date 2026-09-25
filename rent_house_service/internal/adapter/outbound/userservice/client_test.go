package userservice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeerioSst/rent-house-service/config"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

func fake(t *testing.T, validate, user string) (*Client, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/validate", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(validate))
	})
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("user_id") != "42" {
			t.Errorf("user_id = %q", r.URL.Query().Get("user_id"))
		}
		w.Write([]byte(user))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return NewClient(config.Config{
		UserServiceURL: srv.URL + "/", UserServiceValidatePath: "/api/v1/validate", UserServiceUserPath: "/user",
		UserServiceTimeout: time.Second, AuthCacheTTL: time.Minute,
	}), &calls
}

func TestResolve(t *testing.T) {
	c, calls := fake(t, `{"valid":true,"user_id":"42"}`,
		`{"users":{"id":42,"username":"u","password":"HASH","roles":[{"id":1,"name":"admin"}]}}`)
	id, err := c.Resolve(context.Background(), "tok")
	if err != nil || id.UserID != 42 || len(id.Roles) != 1 || id.Roles[0] != "admin" {
		t.Fatalf("got %+v, %v", id, err)
	}
	if _, err := c.Resolve(context.Background(), "tok"); err != nil || calls.Load() != 1 {
		t.Fatalf("second call should hit cache, upstream calls = %d, err = %v", calls.Load(), err)
	}
}

func TestResolveRejects(t *testing.T) {
	c, _ := fake(t, `{"valid":false}`, `{}`)
	if _, err := c.Resolve(context.Background(), "bad"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("invalid token: %v", err)
	}
	c, _ = fake(t, `{"valid":true,"user_id":"42"}`, `{"users":{}}`)
	if _, err := c.Resolve(context.Background(), "tok"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("unknown user: %v", err)
	}
}

func TestResolveUpstreamDown(t *testing.T) {
	c := NewClient(config.Config{UserServiceURL: "http://127.0.0.1:1", UserServiceValidatePath: "/v", UserServiceUserPath: "/u", UserServiceTimeout: time.Second})
	if _, err := c.Resolve(context.Background(), "tok"); !errors.Is(err, domain.ErrUpstreamUnavailable) {
		t.Fatalf("got %v", err)
	}
}
