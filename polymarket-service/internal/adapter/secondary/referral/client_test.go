package referral

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

func client(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(&config.Config{Referral: config.ReferralConfig{BaseURL: srv.URL, Timeout: "2s"}})
}

func TestGenerateLink(t *testing.T) {
	c := client(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path != "/api/v1/referral/generate" || body["owner_user_id"] != "u1" {
			t.Errorf("%s %v", r.URL.Path, body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"ref_code":"ABC","deep_link":"app://open?ref=ABC","expires_at":1893456000000}}`))
	})
	link, err := c.GenerateLink(context.Background(), "u1")
	if err != nil || link.RefCode != "ABC" || link.DeepLink == "" || link.ExpiresAt.Year() != 2030 {
		t.Fatalf("%+v %v", link, err)
	}
}

func TestActivate(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
		owner  string
		want   error
	}{
		"attributed":       {200, `{"data":{"attributed":true,"owner_user_id":"owner"}}`, "owner", nil},
		"not attributed":   {200, `{"data":{"attributed":false}}`, "", port.ErrInvalidInput},
		"self referral":    {400, `{"error":{"code":"SELF_REFERRAL","message":"no"}}`, "", port.ErrReferralNotAllowed},
		"already referred": {400, `{"error":{"code":"ALREADY_REFERRED","message":"no"}}`, "", port.ErrReferralNotAllowed},
		"inactive link":    {422, `{"error":{"code":"LINK_NOT_ACTIVE","message":"used"}}`, "", port.ErrInvalidInput},
		"unknown code":     {404, `{"error":{"code":"NOT_FOUND","message":"nope"}}`, "", port.ErrInvalidInput},
		"outage":           {500, `{"error":{"code":"INTERNAL_ERROR","message":"boom"}}`, "", port.ErrUpstream},
	}
	for name, tc := range cases {
		c := client(t, func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["ref_code"] != "ABC" || body["user_id"] != "u2" {
				t.Errorf("%s: %v", name, body)
			}
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		})
		owner, err := c.Activate(context.Background(), "ABC", "u2")
		if !errors.Is(err, tc.want) || owner != tc.owner {
			t.Errorf("%s: owner %q err %v, want %q / %v", name, owner, err, tc.owner, tc.want)
		}
	}
}

func TestStatsAndUnconfigured(t *testing.T) {
	c := client(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/referral/user/u3/stats" {
			t.Errorf("%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"total_invited":5,"total_installed":3,"total_rewarded":1,"total_reward_amt":12.5}}`))
	})
	s, err := c.Stats(context.Background(), "u3")
	if err != nil || s.TotalInvited != 5 || s.TotalRewardAmt != 12.5 {
		t.Fatalf("%+v %v", s, err)
	}
	if _, err := New(&config.Config{}).Stats(context.Background(), "u3"); !errors.Is(err, port.ErrUpstream) {
		t.Fatalf("%v", err)
	}
}
