package ssohttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type fakeSSO struct {
	gotState string
}

func (f *fakeSSO) LoginURL(_ context.Context, state string) (string, error) {
	f.gotState = state
	return "https://idp.example/auth?state=" + url.QueryEscape(state), nil
}

func (f *fakeSSO) Callback(_ context.Context, code string) (dto.LoginResponse, error) {
	return dto.LoginResponse{SessionToken: "access-" + code, RefreshToken: "refresh", ExpiryTime: 900}, nil
}

func newMux(t *testing.T, f *fakeSSO) *runtime.ServeMux {
	t.Helper()
	mux := runtime.NewServeMux()
	if err := New(f).Register(mux); err != nil {
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
	if len(cookies) != 1 || cookies[0].Name != stateCookie || cookies[0].Value != f.gotState || !cookies[0].HttpOnly {
		t.Fatalf("state cookie: %+v", cookies)
	}

	req := httptest.NewRequest(http.MethodGet, CallbackPath+"?code=abc&state="+url.QueryEscape(f.gotState), nil)
	req.AddCookie(cookies[0])
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK || body["session_token"] != "access-abc" {
		t.Fatalf("callback: %d %v", rec.Code, body)
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
