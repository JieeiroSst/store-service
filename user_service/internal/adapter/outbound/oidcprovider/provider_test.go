package oidcprovider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
)

func fakeIssuer(t *testing.T, claims map[string]any) *httptest.Server {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: "k1"}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var srv *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                srv.URL,
			"authorization_endpoint":                srv.URL + "/authorize",
			"token_endpoint":                        srv.URL + "/token",
			"jwks_uri":                              srv.URL + "/keys",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"}}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code") != "good" {
			http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
			return
		}
		body := map[string]any{"iss": srv.URL, "aud": "client", "sub": "sub-1", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
		for k, v := range claims {
			body[k] = v
		}
		payload, _ := json.Marshal(body)
		obj, err := signer.Sign(payload)
		if err != nil {
			t.Error(err)
			return
		}
		idToken, _ := obj.CompactSerialize()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "at", "token_type": "Bearer", "expires_in": 3600, "id_token": idToken})
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestExchange(t *testing.T) {
	srv := fakeIssuer(t, map[string]any{"email": "quan@example.com", "email_verified": "true", "name": "Quan", "preferred_username": "quan"})
	p := New(Config{Issuer: srv.URL, ClientID: "client", ClientSecret: "secret", RedirectURL: "http://app/callback"})
	ctx := context.Background()

	raw, err := p.AuthCodeURL(ctx, "st")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(raw)
	if !strings.HasPrefix(raw, srv.URL+"/authorize") || u.Query().Get("state") != "st" || u.Query().Get("scope") != "openid profile email" {
		t.Fatalf("auth url: %s", raw)
	}

	id, err := p.Exchange(ctx, "good")
	if err != nil {
		t.Fatal(err)
	}
	if id.Subject != "sub-1" || id.Email != "quan@example.com" || !id.EmailVerified || id.Name != "Quan" || id.Username != "quan" {
		t.Fatalf("%+v", id)
	}

	if _, err := p.Exchange(ctx, "bad"); err == nil {
		t.Fatal("expected exchange error")
	}
}

func TestExchangeRejectsWrongAudience(t *testing.T) {
	srv := fakeIssuer(t, map[string]any{"aud": "someone-else", "email": "quan@example.com", "email_verified": true})
	p := New(Config{Issuer: srv.URL, ClientID: "client"})
	if _, err := p.Exchange(context.Background(), "good"); err == nil {
		t.Fatal("expected audience error")
	}
}
