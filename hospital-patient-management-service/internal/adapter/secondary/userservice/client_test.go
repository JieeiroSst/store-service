package userservice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

func newClient(t *testing.T, h http.HandlerFunc) *Client {
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(&config.Config{UserService: config.UpstreamConfig{BaseURL: srv.URL, Timeout: time.Second}})
}

func TestValidateSendsTheTokenAndReturnsTheUser(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/validate" || body["session_token"] != "tok" {
			t.Errorf("unexpected request %s %s %v", r.Method, r.URL.Path, body)
		}
		_, _ = w.Write([]byte(`{"valid":true,"userId":"17"}`))
	})
	if id, err := c.Validate(context.Background(), "tok"); err != nil || id != "17" {
		t.Fatalf("id=%q err=%v", id, err)
	}
}

func TestValidateClassifiesFailures(t *testing.T) {
	cases := map[string]struct {
		h    http.HandlerFunc
		want error
	}{
		"401":         {func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(401) }, model.ErrUnauthenticated},
		"403":         {func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(403) }, model.ErrUnauthenticated},
		"valid=false": {func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"valid":false}`)) }, model.ErrUnauthenticated},
		"no user id":  {func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"valid":true}`)) }, model.ErrUnauthenticated},
		"500":         {func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(500) }, model.ErrUpstream},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := newClient(t, tc.h).Validate(context.Background(), "t"); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
