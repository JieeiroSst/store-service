package authn

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func jwt(claims map[string]any) string {
	b, _ := json.Marshal(claims)
	return "h." + base64.RawURLEncoding.EncodeToString(b) + ".s"
}

func userService(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			SessionToken string `json:"session_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		switch in.SessionToken {
		case jwt(map[string]any{"username": "an", "role": "user"}):
			_, _ = w.Write([]byte(`{"valid":true,"user_id":"7"}`))
		case jwt(map[string]any{"username": "boss", "role": "admin"}):
			_, _ = w.Write([]byte(`{"valid":true,"user_id":"1"}`))
		case "down":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			_, _ = w.Write([]byte(`{"valid":false}`))
		}
	}))
}

func TestAuthenticate(t *testing.T) {
	us := userService(t)
	defer us.Close()
	ctx := context.Background()
	user := jwt(map[string]any{"username": "an", "role": "user"})
	admin := jwt(map[string]any{"username": "boss", "role": "admin"})

	open := New(us.URL, nil, 0)
	strict := New(us.URL, []string{"s3cret"}, 0)

	p, err := strict.Authenticate(ctx, "Bearer "+user, "")
	if err != nil || p.UserID != "7" || p.IsAdmin() || !p.Owns("7") || p.Owns("8") {
		t.Fatalf("user principal = %+v, %v", p, err)
	}
	if p, err := strict.Authenticate(ctx, "Bearer "+admin, ""); err != nil || !p.IsAdmin() || !p.Owns("8") {
		t.Fatalf("admin principal = %+v, %v", p, err)
	}
	if _, err := strict.Authenticate(ctx, "Bearer bogus", ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("bogus token err = %v", err)
	}
	if _, err := strict.Authenticate(ctx, "Bearer down", ""); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("down err = %v", err)
	}
	if p, err := strict.Authenticate(ctx, "", "s3cret"); err != nil || !p.Service || !p.IsAdmin() {
		t.Fatalf("internal principal = %+v, %v", p, err)
	}
	if _, err := strict.Authenticate(ctx, "", "wrong"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("wrong internal err = %v", err)
	}
	if _, err := strict.Request(httptest.NewRequest(http.MethodGet, "/", nil)); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("anonymous strict err = %v", err)
	}
	if _, err := open.Request(httptest.NewRequest(http.MethodGet, "/", nil)); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("anonymous without internal tokens err = %v", err)
	}
	if _, err := open.Authenticate(ctx, "", "anything"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("internal token without internal tokens err = %v", err)
	}
	if _, err := open.Authenticate(ctx, "Bearer bogus", ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("bogus token without internal tokens err = %v", err)
	}
}

func TestMiddleware(t *testing.T) {
	us := userService(t)
	defer us.Close()
	a := New(us.URL, []string{"s3cret"}, 0)
	h := a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := FromContext(r.Context())
		_, _ = w.Write([]byte(p.UserID))
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+jwt(map[string]any{"username": "an", "role": "user"}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || w.Body.String() != "7" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous got %d", w.Code)
	}
}

func TestOptionalMiddlewareAndRequire(t *testing.T) {
	us := userService(t)
	defer us.Close()
	a := New(us.URL, []string{"s3cret"}, 0)
	h := a.OptionalMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := RequireOwner(r.Context(), "7"); err != nil {
			WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	run := func(set func(*http.Request)) int {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		set(r)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	cases := []struct {
		name string
		set  func(*http.Request)
		want int
	}{
		{"anonymous", func(r *http.Request) {}, http.StatusUnauthorized},
		{"owner", func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+jwt(map[string]any{"username": "an", "role": "user"}))
		}, http.StatusOK},
		{"admin", func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+jwt(map[string]any{"username": "boss", "role": "admin"}))
		}, http.StatusOK},
		{"internal", func(r *http.Request) { r.Header.Set(InternalTokenHeader, "s3cret") }, http.StatusOK},
		{"bad token", func(r *http.Request) { r.Header.Set("Authorization", "Bearer nope") }, http.StatusUnauthorized},
		{"down", func(r *http.Request) { r.Header.Set("Authorization", "Bearer down") }, http.StatusServiceUnavailable},
	}
	for _, c := range cases {
		if got := run(c.set); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
	other := New(us.URL, []string{"s3cret"}, 0).OptionalMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := RequireOwner(r.Context(), "99"); err != nil {
			WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+jwt(map[string]any{"username": "an", "role": "user"}))
	w := httptest.NewRecorder()
	other.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("non-owner got %d", w.Code)
	}
}
