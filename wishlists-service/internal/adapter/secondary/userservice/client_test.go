package userservice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/config"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/pkg/credential"
)

func TestGetUser(t *testing.T) {
	var gotAuth, gotInternal, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotInternal, gotPath = r.Header.Get("Authorization"), r.Header.Get("X-Internal-Token"), r.URL.Path
		if gotAuth != "Bearer access-1" && gotInternal != "svc" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/v1/users/7/profile":
			_, _ = w.Write([]byte(`{"id":7,"username":"an","name":"An Nguyen","active":true}`))
		case "/api/v1/users/8/profile":
			_, _ = w.Write([]byte(`{"id":9,"username":"other"}`))
		case "/api/v1/users/500/profile":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := New(&config.Config{UserService: config.UserServiceConfig{URL: srv.URL + "/", Timeout: time.Second}})
	user := credential.With(context.Background(), credential.Credential{Bearer: "access-1"})

	u, err := c.GetUser(user, "7")
	if err != nil {
		t.Fatal(err)
	}
	if *u != (model.Owner{ID: "7", Username: "an", Name: "An Nguyen", Active: true}) {
		t.Fatalf("unexpected owner %+v", u)
	}
	if gotAuth != "Bearer access-1" || gotInternal != "" || gotPath != "/api/v1/users/7/profile" {
		t.Fatalf("caller access token must be forwarded: auth=%q internal=%q path=%q", gotAuth, gotInternal, gotPath)
	}

	service := credential.With(context.Background(), credential.Credential{Internal: "svc"})
	if _, err := c.GetUser(service, "7"); err != nil || gotInternal != "svc" || gotAuth != "" {
		t.Fatalf("internal caller token must be forwarded: %v auth=%q internal=%q", err, gotAuth, gotInternal)
	}

	if _, err := c.GetUser(context.Background(), "7"); !errors.Is(err, model.ErrUnauthenticated) {
		t.Fatalf("no credentials must not call user-service: %v", err)
	}
	expired := credential.With(context.Background(), credential.Credential{Bearer: "expired"})
	if _, err := c.GetUser(expired, "7"); !errors.Is(err, model.ErrUnauthenticated) {
		t.Fatalf("rejected token: %v", err)
	}
	if _, err := c.GetUser(user, "404"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("missing user: %v", err)
	}
	if _, err := c.GetUser(user, "abc"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("non numeric id: %v", err)
	}
	if _, err := c.GetUser(user, "500"); !errors.Is(err, model.ErrUnavailable) {
		t.Fatalf("server error: %v", err)
	}
	if _, err := c.GetUser(user, "8"); !errors.Is(err, model.ErrUnavailable) {
		t.Fatalf("mismatched id: %v", err)
	}
	srv.Close()
	if _, err := c.GetUser(user, "7"); !errors.Is(err, model.ErrUnavailable) {
		t.Fatalf("connection refused: %v", err)
	}
}
