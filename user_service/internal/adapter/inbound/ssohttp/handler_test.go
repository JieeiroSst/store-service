package ssohttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type fakeSSO struct {
	gotState    string
	gotVerifier string
	cbVerifier  string
	client      domain.ClientInfo
}

func (f *fakeSSO) LoginURL(_ context.Context, state, verifier string) (string, error) {
	f.gotState, f.gotVerifier = state, verifier
	return "https://idp.example/auth?state=" + url.QueryEscape(state), nil
}

func (f *fakeSSO) Callback(ctx context.Context, code, verifier string) (dto.LoginResponse, error) {
	f.cbVerifier = verifier
	f.client = domain.ClientInfoFrom(ctx)
	return dto.LoginResponse{SessionToken: "access-" + code, RefreshToken: "refresh", ExpiryTime: 900}, nil
}

func (f *fakeSSO) LoginWithToken(_ context.Context, idToken string) (dto.LoginResponse, error) {
	if idToken == "" {
		return dto.LoginResponse{}, domain.ErrInvalidRequest
	}
	return dto.LoginResponse{SessionToken: "access-" + idToken}, nil
}

const (
	LoginPath    = SSOBasePath + "/login"
	CallbackPath = SSOBasePath + "/callback"
)

func newMux(t *testing.T, f *fakeSSO) *runtime.ServeMux {
	t.Helper()
	mux := runtime.NewServeMux()
	if err := New(f, SSOBasePath).Register(mux); err != nil {
		t.Fatal(err)
	}
	return mux
}

func TestLoginThenCallback(t *testing.T) {
	f := &fakeSSO{}
	mux := newMux(t, f)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, LoginPath, nil))
	if rec.Code != http.StatusFound || f.gotState == "" {
		t.Fatalf("login: %d state=%q", rec.Code, f.gotState)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != stateCookie || cookies[0].Value != f.gotState+"."+f.gotVerifier || !cookies[0].HttpOnly || len(f.gotVerifier) < 43 {
		t.Fatalf("state cookie: %+v verifier=%q", cookies, f.gotVerifier)
	}

	req := httptest.NewRequest(http.MethodGet, CallbackPath+"?code=abc&state="+url.QueryEscape(f.gotState), nil)
	req.Header.Set("X-Forwarded-For", "9.9.9.9, 10.0.0.1")
	req.Header.Set("User-Agent", "Safari")
	req.AddCookie(cookies[0])
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body["session_token"] != "access-abc" {
		t.Fatalf("callback: %d %v", rec.Code, body)
	}
	if f.cbVerifier != f.gotVerifier || f.client.IP != "9.9.9.9" || f.client.UserAgent != "Safari" {
		t.Fatalf("verifier=%q client=%+v", f.cbVerifier, f.client)
	}
}

func TestCallbackRejectsBadState(t *testing.T) {
	mux := newMux(t, &fakeSSO{})
	cases := map[string]*http.Cookie{
		"no cookie":      nil,
		"wrong cookie":   {Name: stateCookie, Value: "other"},
		"matching empty": {Name: stateCookie, Value: ""},
	}
	for name, c := range cases {
		req := httptest.NewRequest(http.MethodGet, CallbackPath+"?code=abc&state=s1", nil)
		if c != nil {
			req.AddCookie(c)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if rec.Code != http.StatusUnauthorized || body["code"] != "AUTH_SSO_FAILED" {
			t.Fatalf("%s: %d %v", name, rec.Code, body)
		}
	}
}

func TestCallbackProviderError(t *testing.T) {
	mux := newMux(t, &fakeSSO{})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, CallbackPath+"?error=access_denied", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestSocialRoutes(t *testing.T) {
	const GoogleBasePath = SocialBasePath + "google"
	f := &fakeSSO{}
	mux := runtime.NewServeMux()
	if err := New(f, GoogleBasePath).Register(mux); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, GoogleBasePath+"/login", nil))
	cookies := rec.Result().Cookies()
	if rec.Code != http.StatusFound || len(cookies) != 1 || cookies[0].Path != GoogleBasePath {
		t.Fatalf("login: %d %+v", rec.Code, cookies)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, GoogleBasePath+"/token", strings.NewReader(`{"id_token":"idt"}`)))
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body["session_token"] != "access-idt" {
		t.Fatalf("token: %d %v", rec.Code, body)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, GoogleBasePath+"/token", strings.NewReader(`{"access_token":"fb"}`)))
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body["session_token"] != "access-fb" {
		t.Fatalf("access_token: %d %v", rec.Code, body)
	}

	for _, payload := range []string{`not json`, `{}`} {
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, GoogleBasePath+"/token", strings.NewReader(payload)))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%q: %d %s", payload, rec.Code, rec.Body)
		}
	}
}
