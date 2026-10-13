package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/wishlists-service/internal/application"
	"github.com/JIeeiroSst/wishlists-service/internal/config"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"github.com/JIeeiroSst/wishlists-service/pkg/authn"
	"github.com/JIeeiroSst/wishlists-service/pkg/credential"
)

const internalToken = "svc-secret"

func token(user string) string {
	b, _ := json.Marshal(map[string]string{"username": user, "role": "user"})
	return "h." + base64.RawURLEncoding.EncodeToString(b) + ".s"
}

func newServer(t *testing.T) *httptest.Server {
	users := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			SessionToken string `json:"session_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		switch in.SessionToken {
		case token("alice"):
			_, _ = w.Write([]byte(`{"valid":true,"user_id":"1"}`))
		case token("bob"):
			_, _ = w.Write([]byte(`{"valid":true,"user_id":"2"}`))
		default:
			_, _ = w.Write([]byte(`{"valid":false}`))
		}
	}))
	t.Cleanup(users.Close)
	repo := memory.New()
	directory := memory.NewUsers(model.Owner{ID: "1", Username: "alice", Name: "Alice", Active: true}, model.Owner{ID: "2", Username: "bob", Active: true})
	h := NewHandler(Params{
		Wishlists: application.NewWishlistService(repo, directory),
		Health:    []port.HealthChecker{repo, flakyHealth{}},
		Limiter:   &countingLimiter{counts: map[string]int{}},
		Auth:      authn.New(users.URL, []string{internalToken}, time.Second),
		Config:    &config.Config{RateLimit: config.RateLimitConfig{WritesPerMinute: writesPerMinute}},
	})
	srv := httptest.NewServer(h.Routes())
	t.Cleanup(srv.Close)
	return srv
}

const writesPerMinute = 30

type flakyHealth struct{}

func (flakyHealth) Name() string { return "redis" }

func (flakyHealth) Critical() bool { return false }

func (flakyHealth) Ping(context.Context) error { return errors.New("down") }

type countingLimiter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (l *countingLimiter) Allow(_ context.Context, key string, limit int, _ time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts[key]++
	return l.counts[key] <= limit, nil
}

type response struct {
	status int
	body   map[string]any
}

func call(t *testing.T, srv *httptest.Server, method, path, user string, body any) response {
	t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, srv.URL+path, rd)
	switch {
	case user == "internal":
		req.Header.Set(authn.InternalTokenHeader, internalToken)
	case user != "":
		req.Header.Set("Authorization", "Bearer "+token(user))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out := response{status: resp.StatusCode}
	_ = json.NewDecoder(resp.Body).Decode(&out.body)
	return out
}

func data(r response) map[string]any {
	m, _ := r.body["data"].(map[string]any)
	return m
}

func TestHealth(t *testing.T) {
	srv := newServer(t)
	if r := call(t, srv, http.MethodGet, "/livez", "", nil); r.status != http.StatusOK {
		t.Fatalf("livez: %d", r.status)
	}
	r := call(t, srv, http.MethodGet, "/health", "", nil)
	if r.status != http.StatusOK || r.body["status"] != "degraded" {
		t.Fatalf("non critical failure should degrade but stay ready: %d %v", r.status, r.body)
	}
}

func TestWishlistFlow(t *testing.T) {
	srv := newServer(t)

	if r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "", map[string]string{"name": "x"}); r.status != http.StatusUnauthorized {
		t.Fatalf("anonymous create: %d", r.status)
	}

	r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "alice", map[string]string{"name": "Birthday", "visibility": "shared"})
	if r.status != http.StatusCreated {
		t.Fatalf("create: %d %v", r.status, r.body)
	}
	w := data(r)
	id := strconv.Itoa(int(w["id"].(float64)))
	key := w["key"].(string)
	if w["owner_id"] != "1" || w["visibility"] != "Shared" || w["owner"].(map[string]any)["username"] != "alice" {
		t.Fatalf("unexpected wishlist %v", w)
	}

	if r := call(t, srv, http.MethodGet, "/api/v1/wishlists/"+id, "bob", nil); r.status != http.StatusNotFound {
		t.Fatalf("bob get shared by id: %d", r.status)
	}
	if r := call(t, srv, http.MethodGet, "/api/v1/shared-wishlists/"+key, "", nil); r.status != http.StatusOK {
		t.Fatalf("anonymous get by key: %d", r.status)
	}

	r = call(t, srv, http.MethodPost, "/api/v1/wishlists/"+id+"/lines", "alice", map[string]any{"product_id": 42, "title": "Book"})
	if r.status != http.StatusCreated {
		t.Fatalf("add line: %d %v", r.status, r.body)
	}
	lineID := strconv.Itoa(int(data(r)["id"].(float64)))
	if r := call(t, srv, http.MethodPost, "/api/v1/wishlists/"+id+"/lines", "bob", map[string]any{"product_id": 1, "title": "x"}); r.status != http.StatusNotFound {
		t.Fatalf("bob add line: %d", r.status)
	}

	r = call(t, srv, http.MethodPatch, "/api/v1/wishlists/"+id+"/lines/"+lineID, "alice", map[string]int{"quantity": 4})
	if r.status != http.StatusOK || data(r)["quantity"].(float64) != 4 {
		t.Fatalf("update line: %d %v", r.status, r.body)
	}
	if r := call(t, srv, http.MethodPatch, "/api/v1/wishlists/"+id+"/lines/"+lineID, "alice", map[string]int{"quantity": 0}); r.status != http.StatusBadRequest {
		t.Fatalf("zero quantity: %d", r.status)
	}

	r = call(t, srv, http.MethodPut, "/api/v1/wishlists/"+id, "alice", map[string]string{"visibility": "public"})
	if r.status != http.StatusOK || data(r)["visibility"] != "Public" {
		t.Fatalf("update: %d %v", r.status, r.body)
	}
	if r := call(t, srv, http.MethodPut, "/api/v1/wishlists/"+id, "bob", map[string]string{"name": "mine"}); r.status != http.StatusForbidden {
		t.Fatalf("bob update public: %d", r.status)
	}

	r = call(t, srv, http.MethodGet, "/api/v1/users/1/wishlists", "", nil)
	if items, _ := data(r)["items"].([]any); r.status != http.StatusOK || len(items) != 1 || data(r)["total"].(float64) != 1 {
		t.Fatalf("public list: %d %v", r.status, r.body)
	}
	r = call(t, srv, http.MethodGet, "/api/v1/wishlists", "bob", nil)
	if items, _ := data(r)["items"].([]any); r.status != http.StatusOK || len(items) != 0 {
		t.Fatalf("bob own list: %d %v", r.status, r.body)
	}
	r = call(t, srv, http.MethodGet, "/api/v1/wishlists", "internal", nil)
	if items, _ := data(r)["items"].([]any); r.status != http.StatusOK || len(items) != 1 {
		t.Fatalf("internal list all: %d %v", r.status, r.body)
	}

	if r := call(t, srv, http.MethodDelete, "/api/v1/wishlists/"+id+"/lines/"+lineID, "alice", nil); r.status != http.StatusNoContent {
		t.Fatalf("remove line: %d", r.status)
	}
	if r := call(t, srv, http.MethodDelete, "/api/v1/wishlists/"+id, "alice", nil); r.status != http.StatusNoContent {
		t.Fatalf("delete: %d", r.status)
	}
	if r := call(t, srv, http.MethodGet, "/api/v1/wishlists/"+id, "alice", nil); r.status != http.StatusNotFound {
		t.Fatalf("get after delete: %d", r.status)
	}
}

func TestBadInput(t *testing.T) {
	srv := newServer(t)
	if r := call(t, srv, http.MethodGet, "/api/v1/wishlists/abc", "alice", nil); r.status != http.StatusBadRequest {
		t.Fatalf("non-numeric id: %d", r.status)
	}
	if r := call(t, srv, http.MethodGet, "/api/v1/wishlists?limit=x", "alice", nil); r.status != http.StatusBadRequest {
		t.Fatalf("bad limit: %d", r.status)
	}
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/wishlists", strings.NewReader("{"))
	req.Header.Set("Authorization", "Bearer "+token("alice"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed json: %d", resp.StatusCode)
	}
	if r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "nobody", nil); r.status != http.StatusUnauthorized {
		t.Fatalf("invalid token: %d", r.status)
	}
}

func TestMeEndpoints(t *testing.T) {
	srv := newServer(t)
	r := call(t, srv, http.MethodGet, "/api/v1/me/wishlists/default", "alice", nil)
	if r.status != http.StatusOK || data(r)["name"] != "Default" {
		t.Fatalf("default: %d %v", r.status, r.body)
	}
	defaultID := data(r)["id"].(float64)
	r = call(t, srv, http.MethodPost, "/api/v1/me/wishlist-items", "alice", map[string]any{"product_id": 77, "title": "Kettle"})
	if r.status != http.StatusCreated || data(r)["wishlist_id"].(float64) != defaultID {
		t.Fatalf("quick add: %d %v", r.status, r.body)
	}
	r = call(t, srv, http.MethodGet, "/api/v1/me/products/77/wishlists", "alice", nil)
	if r.status != http.StatusOK || data(r)["in_wishlist"] != true {
		t.Fatalf("find product: %d %v", r.status, r.body)
	}
	r = call(t, srv, http.MethodGet, "/api/v1/me/products/77/wishlists", "bob", nil)
	if r.status != http.StatusOK || data(r)["in_wishlist"] != false {
		t.Fatalf("bob find product: %d %v", r.status, r.body)
	}
	if r := call(t, srv, http.MethodGet, "/api/v1/me/wishlists/default", "", nil); r.status != http.StatusUnauthorized {
		t.Fatalf("anonymous default: %d", r.status)
	}
	r = call(t, srv, http.MethodGet, "/api/v1/wishlist-stats/popular-products?limit=5", "", nil)
	if list, _ := r.body["data"].([]any); r.status != http.StatusOK || len(list) != 1 {
		t.Fatalf("popular: %d %v", r.status, r.body)
	}
}

func TestCopyMoveClear(t *testing.T) {
	srv := newServer(t)
	r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "alice", map[string]string{"name": "Party", "visibility": "shared"})
	key := data(r)["key"].(string)
	src := strconv.Itoa(int(data(r)["id"].(float64)))
	r = call(t, srv, http.MethodPost, "/api/v1/wishlists/"+src+"/lines", "alice", map[string]any{"product_id": 1, "title": "Balloon", "quantity": 3})
	line := strconv.Itoa(int(data(r)["id"].(float64)))

	r = call(t, srv, http.MethodPost, "/api/v1/shared-wishlists/"+key+"/copy", "bob", nil)
	if r.status != http.StatusCreated || data(r)["owner_id"] != "2" || len(data(r)["lines"].([]any)) != 1 {
		t.Fatalf("copy: %d %v", r.status, r.body)
	}

	r = call(t, srv, http.MethodPost, "/api/v1/wishlists", "alice", map[string]string{"name": "Later"})
	dst := data(r)["id"].(float64)
	r = call(t, srv, http.MethodPost, "/api/v1/wishlists/"+src+"/lines/"+line+"/move", "alice", map[string]any{"target_wishlist_id": dst})
	if r.status != http.StatusOK || data(r)["wishlist_id"].(float64) != dst {
		t.Fatalf("move: %d %v", r.status, r.body)
	}
	dstPath := "/api/v1/wishlists/" + strconv.Itoa(int(dst)) + "/lines"
	if r := call(t, srv, http.MethodDelete, dstPath, "bob", nil); r.status != http.StatusNotFound {
		t.Fatalf("bob clear: %d", r.status)
	}
	r = call(t, srv, http.MethodDelete, dstPath, "alice", nil)
	if r.status != http.StatusOK || data(r)["removed"].(float64) != 1 {
		t.Fatalf("clear: %d %v", r.status, r.body)
	}
}

func TestWriteRateLimit(t *testing.T) {
	srv := newServer(t)
	for i := 0; i < writesPerMinute; i++ {
		if r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "alice", map[string]string{}); r.status != http.StatusCreated {
			t.Fatalf("write %d: %d", i, r.status)
		}
	}
	if r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "alice", map[string]string{}); r.status != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", r.status)
	}
	if r := call(t, srv, http.MethodGet, "/api/v1/wishlists", "alice", nil); r.status != http.StatusOK {
		t.Fatalf("reads are not limited: %d", r.status)
	}
	if r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "bob", map[string]string{}); r.status != http.StatusCreated {
		t.Fatalf("other users are not affected: %d", r.status)
	}
	if r := call(t, srv, http.MethodPost, "/api/v1/wishlists", "internal", map[string]string{"owner_id": "1"}); r.status != http.StatusCreated {
		t.Fatalf("internal callers are exempt: %d", r.status)
	}
}

func TestRedactPath(t *testing.T) {
	cases := map[string]string{
		"/api/v1/shared-wishlists/AbC123":      "/api/v1/shared-wishlists/{key}",
		"/api/v1/shared-wishlists/AbC123/copy": "/api/v1/shared-wishlists/{key}/copy",
		"/api/v1/wishlists/5":                  "/api/v1/wishlists/5",
	}
	for in, want := range cases {
		if got := redactPath(in); got != want {
			t.Errorf("redactPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestForwardedCredentials(t *testing.T) {
	var seen credential.Credential
	h := forwardCredentials(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = credential.From(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wishlists", nil)
	req.Header.Set("Authorization", "Bearer access-1")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !seen.Empty() {
		t.Fatalf("credentials must only be forwarded for authenticated requests: %+v", seen)
	}
	req = req.WithContext(authn.WithPrincipal(req.Context(), authn.Principal{UserID: "1"}))
	h.ServeHTTP(httptest.NewRecorder(), req)
	if seen.Bearer != "access-1" {
		t.Fatalf("bearer not forwarded: %+v", seen)
	}
}
