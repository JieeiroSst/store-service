package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/auth"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
	"github.com/gin-gonic/gin"
)

func whoami(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.Use(NewAuthenticator(cfg).Middleware())
	e.GET("/me", func(c *gin.Context) {
		p, _ := auth.FromContext(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"sub": p.Subject, "name": p.Name, "roles": p.Roles, "service": p.Service})
	})
	return e
}

func call(e *gin.Engine, headers map[string]string) (int, map[string]any) {
	req := httptest.NewRequest("GET", "/me", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func authConfig(users *testsupport.UserService) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{APIKey: "s3cret"},
		Auth:   config.AuthConfig{APIKeyRole: "viewer", UserServiceURL: users.URL(), RolePrefix: "crm-"},
	}
}

func bearer(tok string) map[string]string { return map[string]string{"Authorization": "Bearer " + tok} }

func TestAuth_ClosedWithoutCredentials(t *testing.T) {
	// There is no "open" mode any more: with nothing configured every
	// request still needs a user-service token.
	if code, _ := call(whoami(&config.Config{}), nil); code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401", code)
	}
}

func TestAuth_APIKey(t *testing.T) {
	e := whoami(&config.Config{Server: config.ServerConfig{APIKey: "s3cret"}, Auth: config.AuthConfig{APIKeyRole: "manager"}})

	code, body := call(e, map[string]string{"X-API-Key": "s3cret"})
	if code != 200 || body["service"] != true || body["roles"].([]any)[0] != "manager" || body["sub"] != "api-key" {
		t.Fatalf("got %d %v", code, body)
	}
	if code, _ := call(e, map[string]string{"X-API-Key": "wrong"}); code != 401 {
		t.Fatalf("wrong key: %d", code)
	}
	if code, _ := call(e, nil); code != 401 {
		t.Fatalf("no credentials: %d", code)
	}
}

func TestAuth_UserServiceToken(t *testing.T) {
	users := testsupport.NewUserService(t)
	e := whoami(authConfig(users))

	cases := []struct {
		role string
		want any // CRM role, nil = none
	}{
		{"crm-staff", "staff"},
		{"crm-manager", "manager"},
		{"crm-viewer", "viewer"},
		{"admin", "admin"},
		{"super_admin", "admin"},
		{"operator", "staff"},
		{"user", nil},
		{"crm-unknown", nil},
	}
	for _, tc := range cases {
		code, body := call(e, bearer(users.Token(t, "42", "nhung", tc.role)))
		if code != 200 || body["sub"] != "42" || body["name"] != "nhung" || body["service"] != false {
			t.Fatalf("%s: got %d %v", tc.role, code, body)
		}
		var got any
		if roles, ok := body["roles"].([]any); ok && len(roles) == 1 {
			got = roles[0]
		}
		if got != tc.want {
			t.Errorf("role %q maps to %v, want %v", tc.role, got, tc.want)
		}
	}

	// Several roles: the highest CRM role wins, whichever is primary.
	_, body := call(e, bearer(users.Token(t, "43", "lan", "operator", "operator", "crm-manager", "crm-staff", "crm-viewer", "user")))
	if roles := body["roles"].([]any); len(roles) != 1 || roles[0] != "manager" {
		t.Errorf("operator + crm-manager: roles = %v, want [manager]", body["roles"])
	}

	// Tokens user-service doesn't recognise (forged, expired, logged out).
	if code, _ := call(e, bearer("e30.eyJzdWIiOiI0MiIsInJvbGUiOiJhZG1pbiJ9.forged")); code != 401 {
		t.Fatalf("forged token: %d", code)
	}
	// The API key still works next to bearer tokens.
	if code, body := call(e, map[string]string{"X-API-Key": "s3cret"}); code != 200 || body["service"] != true {
		t.Fatalf("api key alongside tokens: %d %v", code, body)
	}
}

func TestAuth_UserServiceDown(t *testing.T) {
	users := testsupport.NewUserService(t)
	e := whoami(authConfig(users))
	tok := users.Token(t, "42", "nhung", "crm-staff")
	users.Down = true
	if code, _ := call(e, bearer(tok)); code != 503 {
		t.Fatalf("status %d, want 503 while user-service is unreachable", code)
	}
	// The API key does not depend on it.
	if code, _ := call(e, map[string]string{"X-API-Key": "s3cret"}); code != 200 {
		t.Fatalf("api key while user-service is down: %d", code)
	}
}
