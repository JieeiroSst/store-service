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
	return fakeIssuerWith(t, func(string) map[string]any { return claims }, "", nil)
}

func fakeIssuerWith(t *testing.T, claims func(base string) map[string]any, discoveryIssuerPath string, onToken func(url.Values)) *httptest.Server {
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
			"issuer":                                srv.URL + discoveryIssuerPath,
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
		if onToken != nil {
			onToken(r.Form)
		}
		if r.Form.Get("code") != "good" {
			http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
			return
		}
		body := map[string]any{"iss": srv.URL, "aud": "client", "sub": "sub-1", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
		for k, v := range claims(srv.URL) {
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
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "at", "token_type": "Bearer", "expires_in": 3600, "id_token": idToken, "scope": "openid email profile"})
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestExchange(t *testing.T) {
	srv := fakeIssuer(t, map[string]any{"email": "quan@example.com", "email_verified": "true", "name": "Quan", "preferred_username": "quan", "picture": "https://img.example/q.png"})
	p := New(Config{Issuer: srv.URL, ClientID: "client", ClientSecret: "secret", RedirectURL: "http://app/callback"})
	ctx := context.Background()

	raw, err := p.AuthCodeURL(ctx, "st", "")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(raw)
	if !strings.HasPrefix(raw, srv.URL+"/authorize") || u.Query().Get("state") != "st" || u.Query().Get("scope") != "openid profile email" {
		t.Fatalf("auth url: %s", raw)
	}

	id, err := p.Exchange(ctx, "good", "")
	if err != nil {
		t.Fatal(err)
	}
	if id.Subject != "sub-1" || id.Email != "quan@example.com" || !id.EmailVerified || id.Name != "Quan" || id.Username != "quan" {
		t.Fatalf("%+v", id)
	}
	var profile map[string]any
	if id.AvatarURL != "https://img.example/q.png" || strings.Join(id.Scopes, " ") != "openid email profile" || json.Unmarshal(id.Profile, &profile) != nil || profile["sub"] != "sub-1" {
		t.Fatalf("%+v %s", id, id.Profile)
	}

	if _, err := p.Exchange(ctx, "bad", ""); err == nil {
		t.Fatal("expected exchange error")
	}
}

func TestExchangeRejectsWrongAudience(t *testing.T) {
	srv := fakeIssuer(t, map[string]any{"aud": "someone-else", "email": "quan@example.com", "email_verified": true})
	p := New(Config{Issuer: srv.URL, ClientID: "client"})
	if _, err := p.Exchange(context.Background(), "good", ""); err == nil {
		t.Fatal("expected audience error")
	}
}

func idTokenFrom(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	resp, err := http.PostForm(srv.URL+"/token", url.Values{"code": {"good"}})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body.IDToken
}

func TestVerifyIDTokenAudiences(t *testing.T) {
	srv := fakeIssuer(t, map[string]any{"aud": "android-client", "email": "quan@gmail.com", "email_verified": true, "name": "Quan"})
	raw := idTokenFrom(t, srv)
	ctx := context.Background()

	if _, err := New(Config{Issuer: srv.URL, ClientID: "client"}).VerifyToken(ctx, raw); err == nil {
		t.Fatal("expected audience error")
	}
	id, err := New(Config{Issuer: srv.URL, ClientID: "client", Audiences: []string{"android-client"}}).VerifyToken(ctx, raw)
	if err != nil || id.Email != "quan@gmail.com" || !id.EmailVerified || id.Subject != "sub-1" {
		t.Fatalf("%v %+v", err, id)
	}
	if _, err := New(Config{Issuer: srv.URL, ClientID: "client"}).VerifyToken(ctx, "garbage"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestPKCE(t *testing.T) {
	var gotVerifier string
	claims := func(string) map[string]any { return map[string]any{"email": "quan@gmail.com", "email_verified": true} }
	srv := fakeIssuerWith(t, claims, "", func(f url.Values) { gotVerifier = f.Get("code_verifier") })
	p := New(Config{Issuer: srv.URL, ClientID: "client", PKCE: true})
	ctx := context.Background()

	raw, err := p.AuthCodeURL(ctx, "st", "verifier-123")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(raw)
	if u.Query().Get("code_challenge_method") != "S256" || u.Query().Get("code_challenge") == "" {
		t.Fatalf("auth url: %s", raw)
	}
	if _, err := p.Exchange(ctx, "good", "verifier-123"); err != nil || gotVerifier != "verifier-123" {
		t.Fatalf("%v verifier=%q", err, gotVerifier)
	}
}

func TestMicrosoftMultiTenant(t *testing.T) {
	cases := []struct {
		name         string
		claims       func(base string) map[string]any
		wantErr      bool
		wantVerified bool
	}{
		{"work account, email not verified", func(b string) map[string]any {
			return map[string]any{"iss": b + "/tenant-1/v2.0", "tid": "tenant-1", "email": "quan@contoso.com"}
		}, false, false},
		{"work account, domain owner verified", func(b string) map[string]any {
			return map[string]any{"iss": b + "/tenant-1/v2.0", "tid": "tenant-1", "email": "quan@contoso.com", "xms_edov": true}
		}, false, true},
		{"personal account", func(b string) map[string]any {
			return map[string]any{"iss": b + "/" + microsoftConsumersTenant + "/v2.0", "tid": microsoftConsumersTenant, "email": "quan@outlook.com"}
		}, false, true},
		{"issuer from another tenant", func(b string) map[string]any {
			return map[string]any{"iss": b + "/tenant-2/v2.0", "tid": "tenant-1", "email": "quan@contoso.com"}
		}, true, false},
		{"missing tid", func(b string) map[string]any {
			return map[string]any{"iss": b + "/{tenantid}/v2.0", "email": "quan@contoso.com"}
		}, true, false},
	}
	for _, c := range cases {
		srv := fakeIssuerWith(t, c.claims, "/{tenantid}/v2.0", nil)
		p := New(Config{Issuer: srv.URL, IssuerTemplate: srv.URL + "/{tenantid}/v2.0", ClientID: "client", VerifiedEmail: MicrosoftVerifiedEmail})
		id, err := p.Exchange(context.Background(), "good", "")
		if (err != nil) != c.wantErr || (err == nil && id.EmailVerified != c.wantVerified) {
			t.Fatalf("%s: err=%v identity=%+v", c.name, err, id)
		}
	}
}

func TestMicrosoftIssuer(t *testing.T) {
	if iss, tpl := MicrosoftIssuer(""); iss != "https://login.microsoftonline.com/common/v2.0" || tpl != MicrosoftIssuerTemplate {
		t.Fatal(iss, tpl)
	}
	if iss, tpl := MicrosoftIssuer("72f988bf-86f1-41af-91ab-2d7cd011db47"); iss != "https://login.microsoftonline.com/72f988bf-86f1-41af-91ab-2d7cd011db47/v2.0" || tpl != "" {
		t.Fatal(iss, tpl)
	}
}
