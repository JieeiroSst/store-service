package userservice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
)

func TestGetUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/internal/v1/users/7":
			w.Write([]byte(`{"id":7,"username":"quanluu","name":"Quan Luu","email":"q@mail.com","active":true}`))
		case "/internal/v1/users/8":
			w.Write([]byte(`{"id":9}`))
		case "/internal/v1/users/404":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadGateway)
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL+"/", "tok", time.Second)
	ctx := context.Background()

	u, err := c.GetUser(ctx, 7)
	if err != nil || u.Username != "quanluu" || u.Name != "Quan Luu" || !u.Active {
		t.Fatalf("%v %+v", err, u)
	}
	if _, err := c.GetUser(ctx, 404); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("404: %v", err)
	}
	if _, err := c.GetUser(ctx, 8); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("mismatched id: %v", err)
	}
	if _, err := c.GetUser(ctx, 5); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("502: %v", err)
	}
	if _, err := NewClient(srv.URL, "wrong", time.Second).GetUser(ctx, 7); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("401: %v", err)
	}
}
