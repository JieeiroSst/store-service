package oauthprovider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func fakeGraph(t *testing.T, appID string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code") != "good" || r.Form.Get("client_secret") != "secret" {
			http.Error(w, `{"error":{"message":"bad"}}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{"access_token": "user-token", "token_type": "bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/debug_token", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("access_token") != "app|secret" {
			http.Error(w, "bad app token", http.StatusBadRequest)
			return
		}
		switch q.Get("input_token") {
		case "user-token":
			writeJSON(w, map[string]any{"data": map[string]any{"app_id": appID, "is_valid": true, "user_id": "fb-1", "scopes": []string{"email", "public_profile"}}})
		case "other-app-token":
			writeJSON(w, map[string]any{"data": map[string]any{"app_id": "someone-else", "is_valid": true, "user_id": "fb-1"}})
		default:
			writeJSON(w, map[string]any{"data": map[string]any{"is_valid": false}})
		}
	})
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		mac := hmac.New(sha256.New, []byte("secret"))
		mac.Write([]byte(q.Get("access_token")))
		if q.Get("appsecret_proof") != hex.EncodeToString(mac.Sum(nil)) {
			http.Error(w, "bad proof", http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{"id": "fb-1", "name": "Quan", "email": "quan@example.com", "picture": map[string]any{"data": map[string]any{"url": "https://fb.example/q.jpg"}}})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestFacebook(t *testing.T) {
	srv := fakeGraph(t, "app")
	fb := NewFacebook(Config{ClientID: "app", ClientSecret: "secret", RedirectURL: "http://app/cb", AuthURL: "https://fb.example/dialog", APIURL: srv.URL})
	ctx := context.Background()

	raw, err := fb.AuthCodeURL(ctx, "st", "ignored")
	u, _ := url.Parse(raw)
	if err != nil || u.Query().Get("state") != "st" || u.Query().Get("scope") != "public_profile email" || u.Query().Get("code_challenge") != "" {
		t.Fatalf("%v %s", err, raw)
	}

	id, err := fb.Exchange(ctx, "good", "")
	if err != nil || id.Subject != "fb-1" || id.Email != "quan@example.com" || !id.EmailVerified || id.Name != "Quan" {
		t.Fatalf("%v %+v", err, id)
	}
	if id.AvatarURL != "https://fb.example/q.jpg" || len(id.Scopes) != 2 || !json.Valid(id.Profile) {
		t.Fatalf("%+v %s", id, id.Profile)
	}
	if _, err := fb.Exchange(ctx, "bad", ""); err == nil {
		t.Fatal("expected exchange error")
	}

	if id, err := fb.VerifyToken(ctx, "user-token"); err != nil || id.Subject != "fb-1" {
		t.Fatalf("%v %+v", err, id)
	}
	for _, tok := range []string{"other-app-token", "expired"} {
		if _, err := fb.VerifyToken(ctx, tok); err == nil {
			t.Fatalf("%s: expected rejection", tok)
		}
	}
}

func fakeX(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		user, pass, ok := r.BasicAuth()
		if !ok || user != "client" || pass != "secret" || r.Form.Get("code") != "good" || r.Form.Get("code_verifier") != "verifier-123" {
			http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{"access_token": "x-token", "token_type": "bearer", "expires_in": 7200, "scope": "users.read tweet.read"})
	})
	mux.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer x-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]any{"data": map[string]any{"id": "x-1", "name": "Quan", "username": "quanluu", "profile_image_url": "https://x.example/q.jpg"}})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestTwitter(t *testing.T) {
	srv := fakeX(t)
	x := NewTwitter(Config{ClientID: "client", ClientSecret: "secret", RedirectURL: "http://app/cb", AuthURL: "https://x.example/authorize", APIURL: srv.URL})
	ctx := context.Background()

	if _, err := x.AuthCodeURL(ctx, "st", ""); err == nil {
		t.Fatal("expected pkce verifier error")
	}
	raw, err := x.AuthCodeURL(ctx, "st", "verifier-123")
	u, _ := url.Parse(raw)
	if err != nil || u.Query().Get("code_challenge_method") != "S256" || u.Query().Get("scope") != "users.read tweet.read users.email" {
		t.Fatalf("%v %s", err, raw)
	}

	id, err := x.Exchange(ctx, "good", "verifier-123")
	if err != nil || id.Subject != "x-1" || id.Email != "" || id.EmailVerified || id.Username != "quanluu" {
		t.Fatalf("%v %+v", err, id)
	}
	if id.AvatarURL != "https://x.example/q.jpg" || len(id.Scopes) != 2 || id.Scopes[1] != "tweet.read" || !json.Valid(id.Profile) {
		t.Fatalf("%+v", id)
	}
	if _, err := x.Exchange(ctx, "good", "wrong"); err == nil {
		t.Fatal("expected verifier mismatch")
	}
	if _, err := x.VerifyToken(ctx, "x-token"); !errors.Is(err, domain.ErrSSOTokenUnsupported) {
		t.Fatalf("%v", err)
	}
}
