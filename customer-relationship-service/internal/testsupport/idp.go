package testsupport

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

	"github.com/golang-jwt/jwt/v5"
)

// IDP is a throwaway OpenID Connect provider: a JWKS endpoint and a way to
// mint tokens signed by it.
type IDP struct {
	Issuer   string
	Audience string
	Server   *httptest.Server
	key      *rsa.PrivateKey
}

func NewIDP(t testing.TB, issuer, audience string) *IDP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	p := &IDP{Issuer: issuer, Audience: audience, key: key}
	p.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())
		n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{
			{"kty": "RSA", "kid": "k1", "use": "sig", "alg": "RS256", "n": n, "e": e},
		}})
	}))
	t.Cleanup(p.Server.Close)
	return p
}

// JWKSURL is where the provider publishes its keys.
func (p *IDP) JWKSURL() string { return p.Server.URL }

// Token mints a valid access token for a user with the given realm roles.
func (p *IDP) Token(t testing.TB, subject, name string, roles ...string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": p.Issuer, "aud": p.Audience, "sub": subject, "preferred_username": name,
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
		"realm_access": map[string]any{"roles": roles},
	})
	tok.Header["kid"] = "k1"
	s, err := tok.SignedString(p.key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
