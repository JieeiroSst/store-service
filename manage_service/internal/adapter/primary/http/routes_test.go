package http

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/manage-service/internal/domain/port"
	"github.com/Nerzal/gocloak/v13"
)

type fakeKeycloak struct {
	port.KeycloakUsecase
	adminErr  error
	lastToken string
	lastRealm string
	lastUser  gocloak.User
}

func (f *fakeKeycloak) AdminToken(context.Context) (string, error) {
	return "admin-token", f.adminErr
}

func (f *fakeKeycloak) GetUserByID(_ context.Context, token, realm, userID string) (*gocloak.User, error) {
	f.lastToken, f.lastRealm = token, realm
	if userID == "missing" {
		return nil, &gocloak.APIError{Code: http.StatusNotFound, Message: "user not found"}
	}
	return &gocloak.User{ID: gocloak.StringP(userID)}, nil
}

func (f *fakeKeycloak) CreateUserRepresentation(_ context.Context, token, realm string, user gocloak.User) (string, error) {
	f.lastToken, f.lastRealm, f.lastUser = token, realm, user
	return "new-id", nil
}

func (f *fakeKeycloak) GetUsers(_ context.Context, token, realm string, params gocloak.GetUsersParams) ([]*gocloak.User, error) {
	f.lastToken, f.lastRealm = token, realm
	if params.Max == nil || *params.Max != 5 || params.Enabled == nil || !*params.Enabled {
		return nil, &gocloak.APIError{Code: http.StatusBadRequest, Message: "params not parsed"}
	}
	return []*gocloak.User{}, nil
}

func newTestServer(t *testing.T, kc *fakeKeycloak, key string) *httptest.Server {
	t.Helper()
	router := NewRouter(NewHandler(nil, kc), RouterConfig{AuthorizeKey: key, Timeout: time.Second})
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path, body string, headers map[string]string) (int, string) {
	t.Helper()
	req, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, strings.TrimSpace(string(b))
}

func TestHealthBypassesAPIKey(t *testing.T) {
	srv := newTestServer(t, &fakeKeycloak{}, "secret")
	if code, _ := call(t, srv, "GET", "/health", "", nil); code != http.StatusOK {
		t.Fatalf("want 200, got %d", code)
	}
}

func TestAPIKey(t *testing.T) {
	srv := newTestServer(t, &fakeKeycloak{}, "secret")
	path := "/api/v1/admin/realms/demo/users/u1"
	if code, _ := call(t, srv, "GET", path, "", nil); code != http.StatusForbidden {
		t.Fatalf("want 403 without key, got %d", code)
	}
	key := base64.StdEncoding.EncodeToString([]byte("secret"))
	if code, _ := call(t, srv, "GET", path, "", map[string]string{apiKeyHeader: key}); code != http.StatusOK {
		t.Fatalf("want 200 with key, got %d", code)
	}
}

func TestAdminTokenResolution(t *testing.T) {
	kc := &fakeKeycloak{}
	srv := newTestServer(t, kc, "")

	call(t, srv, "GET", "/api/v1/admin/realms/demo/users/u1", "", nil)
	if kc.lastToken != "admin-token" || kc.lastRealm != "demo" {
		t.Fatalf("want admin fallback token, got token=%s realm=%s", kc.lastToken, kc.lastRealm)
	}

	call(t, srv, "GET", "/api/v1/admin/realms/demo/users/u1", "", map[string]string{"Authorization": "Bearer caller"})
	if kc.lastToken != "caller" {
		t.Fatalf("want caller token, got %s", kc.lastToken)
	}
}

func TestErrorMapping(t *testing.T) {
	kc := &fakeKeycloak{}
	srv := newTestServer(t, kc, "")

	if code, _ := call(t, srv, "GET", "/api/v1/admin/realms/demo/users/missing", "", nil); code != http.StatusNotFound {
		t.Fatalf("want 404 from keycloak error, got %d", code)
	}
	if code, _ := call(t, srv, "POST", "/api/v1/admin/realms/demo/users", "not json", nil); code != http.StatusBadRequest {
		t.Fatalf("want 400 for bad body, got %d", code)
	}

	kc.adminErr = port.ErrAdminNotConfigured
	if code, _ := call(t, srv, "GET", "/api/v1/admin/realms/demo/users/u1", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("want 401 without credentials, got %d", code)
	}
}

func TestCreateUserAndQueryParams(t *testing.T) {
	kc := &fakeKeycloak{}
	srv := newTestServer(t, kc, "")

	code, body := call(t, srv, "POST", "/api/v1/admin/realms/demo/users", `{"username":"bob","attributes":{"team":["a"]}}`, nil)
	if code != http.StatusCreated || body != `{"id":"new-id"}` {
		t.Fatalf("unexpected create response %d %s", code, body)
	}
	if kc.lastUser.Username == nil || *kc.lastUser.Username != "bob" || kc.lastUser.Attributes == nil {
		t.Fatalf("user not decoded: %+v", kc.lastUser)
	}

	if code, body := call(t, srv, "GET", "/api/v1/admin/realms/demo/users?max=5&enabled=true", "", nil); code != http.StatusOK {
		t.Fatalf("query params not parsed: %d %s", code, body)
	}
}
