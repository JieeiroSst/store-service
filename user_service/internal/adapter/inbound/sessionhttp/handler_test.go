package sessionhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type fakeSessions struct {
	token          string
	revoked        string
	includeCurrent bool
}

func (f *fakeSessions) check(token string) error {
	if token != "good" {
		return domain.ErrFailedToken
	}
	f.token = token
	return nil
}

func (f *fakeSessions) List(_ context.Context, token string) (dto.ListSessionsResponse, error) {
	if err := f.check(token); err != nil {
		return dto.ListSessionsResponse{}, err
	}
	return dto.ListSessionsResponse{Sessions: []dto.SessionInfo{{ID: "s1", Provider: "google", Current: true}}}, nil
}

func (f *fakeSessions) Revoke(_ context.Context, token, id string) error {
	if err := f.check(token); err != nil {
		return err
	}
	if id != "s1" {
		return domain.ErrSessionNotFound
	}
	f.revoked = id
	return nil
}

func (f *fakeSessions) RevokeOthers(_ context.Context, token string, includeCurrent bool) (dto.RevokeSessionsResponse, error) {
	if err := f.check(token); err != nil {
		return dto.RevokeSessionsResponse{}, err
	}
	f.includeCurrent = includeCurrent
	return dto.RevokeSessionsResponse{Revoked: 3}, nil
}

func serve(t *testing.T, f *fakeSessions, method, target, auth string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	mux := runtime.NewServeMux()
	if err := New(f).Register(mux); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, target, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec, body
}

func TestList(t *testing.T) {
	rec, body := serve(t, &fakeSessions{}, http.MethodGet, SessionsPath, "Bearer good")
	sessions, _ := body["sessions"].([]any)
	if rec.Code != http.StatusOK || len(sessions) != 1 {
		t.Fatalf("%d %v", rec.Code, body)
	}
}

func TestRequiresBearer(t *testing.T) {
	for _, auth := range []string{"", "good", "Basic good", "Bearer bad"} {
		rec, body := serve(t, &fakeSessions{}, http.MethodGet, SessionsPath, auth)
		if rec.Code != http.StatusUnauthorized || body["code"] != string(domain.CodeTokenInvalid) {
			t.Fatalf("%q: %d %v", auth, rec.Code, body)
		}
	}
}

func TestRevoke(t *testing.T) {
	f := &fakeSessions{}
	rec, _ := serve(t, f, http.MethodDelete, "/api/v1/sessions/s1", "Bearer good")
	if rec.Code != http.StatusNoContent || f.revoked != "s1" {
		t.Fatalf("%d %q", rec.Code, f.revoked)
	}
	rec, body := serve(t, f, http.MethodDelete, "/api/v1/sessions/other", "Bearer good")
	if rec.Code != http.StatusNotFound || body["code"] != string(domain.CodeSessionNotFound) {
		t.Fatalf("%d %v", rec.Code, body)
	}
}

func TestRevokeAll(t *testing.T) {
	f := &fakeSessions{}
	rec, body := serve(t, f, http.MethodDelete, SessionsPath+"?include_current=true", "bearer good")
	if rec.Code != http.StatusOK || body["revoked"] != float64(3) || !f.includeCurrent {
		t.Fatalf("%d %v %v", rec.Code, body, f.includeCurrent)
	}
}
