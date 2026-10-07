package userclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/ekyc-service/config"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

func TestGetUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/internal/v1/users/7":
			w.Write([]byte(`{"id":7,"username":"anguyen","active":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	cfg := &config.Config{UserService: config.UserServiceConfig{BaseURL: srv.URL + "/", Token: "tok", Timeout: "1s"}}
	c := NewUserClient(cfg)
	u, err := c.GetUser(context.Background(), "7")
	if err != nil || u.ID != "7" || u.Username != "anguyen" {
		t.Fatalf("%v %+v", err, u)
	}
	if _, err := c.GetUser(context.Background(), "8"); !errors.Is(err, port.ErrUserNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if _, err := c.GetUser(context.Background(), "../x"); !errors.Is(err, port.ErrUserNotFound) {
		t.Fatalf("non-numeric id: %v", err)
	}
}

func TestValidateSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/validate" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["session_token"] == "good" {
			w.Write([]byte(`{"valid":true,"userId":"7"}`))
			return
		}
		w.Write([]byte(`{"valid":false,"userId":""}`))
	}))
	defer srv.Close()
	v := NewSessionValidator(&config.Config{UserService: config.UserServiceConfig{BaseURL: srv.URL, Timeout: "1s"}})
	if id, err := v.ValidateSession(context.Background(), "good"); err != nil || id != "7" {
		t.Fatalf("%q %v", id, err)
	}
	if _, err := v.ValidateSession(context.Background(), "bad"); !errors.Is(err, port.ErrUnauthenticated) {
		t.Fatalf("bad: %v", err)
	}
}
