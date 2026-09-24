package http

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const issuer = "https://sso.example/auth/realms/crm"

type idp struct {
	key    *rsa.PrivateKey
	server *httptest.Server
	down   bool
}

func newIDP(t *testing.T) *idp {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	p := &idp{key: key}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.down {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())
		n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{
			{"kty": "RSA", "kid": "k1", "use": "sig", "alg": "RS256", "n": n, "e": e},
		}})
	}))
	t.Cleanup(p.server.Close)
	return p
}

func (p *idp) token(t *testing.T, mutate func(jwt.MapClaims), method jwt.SigningMethod, key any) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss": issuer, "sub": "user-1", "preferred_username": "nhung", "aud": "crm-api",
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
		"realm_access": map[string]any{"roles": []string{"crm-staff", "offline_access", "other-app-admin"}},
	}
	if mutate != nil {
		mutate(claims)
	}
	tok := jwt.NewWithClaims(method, claims)
	tok.Header["kid"] = "k1"
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func (p *idp) valid(t *testing.T, mutate func(jwt.MapClaims)) string {
	return p.token(t, mutate, jwt.SigningMethodRS256, p.key)
}

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

func oidcConfig(p *idp) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{APIKey: "s3cret"},
		Auth: config.AuthConfig{
			APIKeyRole: "viewer", OIDCIssuer: issuer, OIDCJWKSURL: p.server.URL, OIDCAudience: "crm-api",
			RolesClaim: "realm_access.roles", RolePrefix: "crm-",
		},
	}
}

func TestAuth_OpenWhenNothingIsConfigured(t *testing.T) {
	code, body := call(whoami(&config.Config{}), nil)
	if code != 200 || body["sub"] != "anonymous" || body["roles"].([]any)[0] != "admin" {
		t.Fatalf("got %d %v", code, body)
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

func TestAuth_BearerToken(t *testing.T) {
	p := newIDP(t)
	e := whoami(oidcConfig(p))
	bearer := func(tok string) map[string]string { return map[string]string{"Authorization": "Bearer " + tok} }

	// A valid token: the caller is the user, and only our prefixed roles count.
	code, body := call(e, bearer(p.valid(t, nil)))
	roles := body["roles"].([]any)
	if code != 200 || body["sub"] != "user-1" || body["name"] != "nhung" || body["service"] != false || len(roles) != 1 || roles[0] != "staff" {
		t.Fatalf("got %d %v", code, body)
	}

	// Several of our roles: all are kept (the highest wins when checking).
	_, body = call(e, bearer(p.valid(t, func(c jwt.MapClaims) {
		c["realm_access"] = map[string]any{"roles": []string{"CRM-Viewer", "crm-manager"}}
	})))
	if len(body["roles"].([]any)) != 2 {
		t.Fatalf("roles = %v", body["roles"])
	}

	// Authenticated but with none of our roles: allowed in, with no rights.
	code, body = call(e, bearer(p.valid(t, func(c jwt.MapClaims) { c["realm_access"] = map[string]any{"roles": []string{"unrelated"}} })))
	if code != 200 || body["roles"] != nil {
		t.Fatalf("got %d %v", code, body)
	}

	hs := jwt.SigningMethodHS256
	for name, tok := range map[string]string{
		"expired":        p.valid(t, func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Hour).Unix() }),
		"no expiry":      p.valid(t, func(c jwt.MapClaims) { delete(c, "exp") }),
		"wrong issuer":   p.valid(t, func(c jwt.MapClaims) { c["iss"] = "https://evil.example" }),
		"wrong audience": p.valid(t, func(c jwt.MapClaims) { c["aud"] = "someone-else" }),
		"no subject":     p.valid(t, func(c jwt.MapClaims) { delete(c, "sub") }),
		"unknown signer": func() string {
			k, _ := rsa.GenerateKey(rand.Reader, 2048)
			return p.token(t, nil, jwt.SigningMethodRS256, k)
		}(),
		"HS256 with the public key as the secret": p.token(t, nil, hs, p.key.N.Bytes()),
		"alg none": p.token(t, nil, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType),
		"garbage":  "not.a.jwt",
	} {
		if code, _ := call(e, bearer(tok)); code != 401 {
			t.Errorf("%s: status %d, want 401", name, code)
		}
	}

	// A tampered payload.
	good := p.valid(t, nil)
	if code, _ := call(e, bearer(good[:len(good)-4]+"AAAA")); code != 401 {
		t.Errorf("tampered signature: %d", code)
	}

	// The API key still works next to OIDC.
	if code, body := call(e, map[string]string{"X-API-Key": "s3cret"}); code != 200 || body["roles"].([]any)[0] != "viewer" {
		t.Errorf("api key alongside oidc: %d %v", code, body)
	}
}

func TestAuth_IdentityProviderDown(t *testing.T) {
	p := newIDP(t)
	p.down = true
	e := whoami(oidcConfig(p))
	if code, _ := call(e, map[string]string{"Authorization": "Bearer " + p.valid(t, nil)}); code != 503 {
		t.Fatalf("status %d, want 503 while the JWKS is unreachable", code)
	}
	// The API key does not depend on it.
	if code, _ := call(e, map[string]string{"X-API-Key": "s3cret"}); code != 200 {
		t.Fatalf("api key while the idp is down: %d", code)
	}
}

func TestClaimStrings(t *testing.T) {
	claims := map[string]any{
		"realm_access":    map[string]any{"roles": []any{"a", "b", 3}},
		"role":            "solo",
		"resource_access": map[string]any{"crm": map[string]any{"roles": []any{"crm-admin"}}},
	}
	for path, want := range map[string]int{"realm_access.roles": 2, "role": 1, "resource_access.crm.roles": 1, "missing": 0, "realm_access.roles.deeper": 0} {
		if got := claimStrings(claims, path); len(got) != want {
			t.Errorf("%s: %v, want %d entries", path, got, want)
		}
	}
}
