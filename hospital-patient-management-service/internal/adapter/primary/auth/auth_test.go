package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type stubValidator struct {
	calls int
	err   error
}

func (s *stubValidator) Validate(_ context.Context, token string) (string, error) {
	s.calls++
	if s.err != nil {
		return "", s.err
	}
	if token != "good" {
		return "", model.ErrUnauthenticated
	}
	return "u1", nil
}

func newAuth(mode string, v *stubValidator) *Authenticator {
	return New(&config.Config{Auth: config.AuthConfig{Mode: mode, CacheTTL: 5 * time.Second}}, v)
}

func do(h http.Handler, path, header string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHTTPRequiresAValidBearerToken(t *testing.T) {
	v := &stubValidator{}
	var seen string
	h := newAuth(config.AuthToken, v).HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = UserID(r.Context())
	}))

	for name, header := range map[string]string{"missing": "", "wrong scheme": "Basic abc", "empty": "Bearer ", "bad token": "Bearer nope"} {
		if rec := do(h, "/v1/patients", header); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want 401", name, rec.Code)
		}
	}
	if rec := do(h, "/v1/patients", "Bearer good"); rec.Code != http.StatusOK || seen != "u1" {
		t.Errorf("valid token: status %d user %q", rec.Code, seen)
	}
}

func TestHealthNeedsNoToken(t *testing.T) {
	h := newAuth(config.AuthToken, &stubValidator{}).HTTP(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if rec := do(h, "/health", ""); rec.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rec.Code)
	}
}

func TestUserServiceOutageIs503NotAnAuthFailure(t *testing.T) {
	h := newAuth(config.AuthToken, &stubValidator{err: model.ErrUpstream}).HTTP(http.NotFoundHandler())
	if rec := do(h, "/v1/patients", "Bearer good"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status %d, want 503", rec.Code)
	}
}

func TestOffModeLetsEverythingThrough(t *testing.T) {
	v := &stubValidator{}
	h := newAuth(config.AuthOff, v).HTTP(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if rec := do(h, "/v1/patients", ""); rec.Code != http.StatusOK || v.calls != 0 {
		t.Errorf("status %d, validator calls %d", rec.Code, v.calls)
	}
}

func TestValidatedTokensAreCachedBriefly(t *testing.T) {
	v := &stubValidator{}
	a := newAuth(config.AuthToken, v)
	clock := time.Now()
	a.now = func() time.Time { return clock }

	for i := 0; i < 3; i++ {
		if _, err := a.Authenticate(context.Background(), "Bearer good"); err != nil {
			t.Fatal(err)
		}
	}
	if v.calls != 1 {
		t.Errorf("validator called %d times, want 1", v.calls)
	}
	clock = clock.Add(6 * time.Second)
	_, _ = a.Authenticate(context.Background(), "Bearer good")
	if v.calls != 2 {
		t.Errorf("after expiry: %d calls, want 2", v.calls)
	}

	// Failures are never cached: a fixed outage must not linger.
	v.err = errors.New("boom")
	_, _ = a.Authenticate(context.Background(), "Bearer other")
	v.err = nil
	if _, err := a.Authenticate(context.Background(), "Bearer good"); err != nil {
		t.Errorf("after recovery: %v", err)
	}
}

func TestZeroTTLChecksEveryRequest(t *testing.T) {
	v := &stubValidator{}
	a := New(&config.Config{Auth: config.AuthConfig{Mode: config.AuthToken}}, v)
	for i := 0; i < 3; i++ {
		_, _ = a.Authenticate(context.Background(), "Bearer good")
	}
	if v.calls != 3 {
		t.Errorf("validator called %d times, want 3", v.calls)
	}
}

func TestCallersTokenIsAvailableToDownstreamCalls(t *testing.T) {
	var got string
	h := newAuth(config.AuthToken, &stubValidator{}).HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = port.BearerFrom(r.Context())
	}))
	do(h, "/v1/patients", "Bearer good")
	if got != "good" {
		t.Errorf("token in context = %q, want good", got)
	}
}
