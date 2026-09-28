package userservice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func fakeToken(claims map[string]any) string {
	raw, _ := json.Marshal(claims)
	return "e30." + base64.RawURLEncoding.EncodeToString(raw) + ".sig"
}

func newFakeUserService(t *testing.T, calls *int32) *httptest.Server {
	t.Helper()
	valid := fakeToken(map[string]any{"sub": "42", "username": "alice", "role": "admin", "roles": []string{"admin", "crm-manager"}})
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/validate":
			atomic.AddInt32(calls, 1)
			var in struct {
				SessionToken string `json:"session_token"`
			}
			_ = json.NewDecoder(r.Body).Decode(&in)
			if in.SessionToken == valid {
				_, _ = w.Write([]byte(`{"valid":true,"userId":"42"}`))
				return
			}
			_, _ = w.Write([]byte(`{}`))
		case "/user":
			if r.URL.Query().Get("user_id") != "42" {
				_, _ = w.Write([]byte(`{"users":{}}`))
				return
			}
			_, _ = w.Write([]byte(`{"users":{"id":42,"username":"alice","email":"a@x.io","checked":true}}`))
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestAuthenticate(t *testing.T) {
	var calls int32
	srv := newFakeUserService(t, &calls)
	defer srv.Close()
	c := New(srv.URL, 0)
	ctx := context.Background()

	token := fakeToken(map[string]any{"sub": "42", "username": "alice", "role": "admin", "roles": []string{"admin", "crm-manager"}})
	id, err := c.Authenticate(ctx, "Bearer "+token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if id.UserID != "42" || id.Username != "alice" || !id.IsAdmin() {
		t.Errorf("identity = %+v", id)
	}
	if !id.HasRole("crm-manager") || !id.HasRole("admin") || id.HasRole("crm-viewer") {
		t.Errorf("roles = %v", id.Roles)
	}

	// Cached: no second round trip.
	if _, err := c.Authenticate(ctx, token); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("validate called %d times, want 1", calls)
	}

	if _, err := c.Authenticate(ctx, "garbage"); !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("invalid token: got %v", err)
	}
	if _, err := c.Authenticate(ctx, ""); !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("empty token: got %v", err)
	}
}

func TestAuthenticateUpstreamDown(t *testing.T) {
	c := New("http://127.0.0.1:1", 0)
	if _, err := c.Authenticate(context.Background(), "x.y.z"); !errors.Is(err, ErrUpstream) {
		t.Errorf("got %v, want ErrUpstream", err)
	}
}

func TestGetUser(t *testing.T) {
	var calls int32
	srv := newFakeUserService(t, &calls)
	defer srv.Close()
	c := New(srv.URL, 0)

	u, err := c.GetUser(context.Background(), "42")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != "42" || u.Username != "alice" || !u.Active {
		t.Errorf("user = %+v", u)
	}
	if _, err := c.GetUser(context.Background(), "7"); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("missing user: got %v", err)
	}
}
