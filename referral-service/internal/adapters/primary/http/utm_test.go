package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	appconfig "github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/domain"
)

func TestBuildShareURL(t *testing.T) {
	utm := shareUTM(
		domain.UTM{Source: "copy", Medium: "referral", Campaign: "referral_program"},
		domain.ChannelZalo,
		domain.UTM{Campaign: "Sale_Tet_2026", Content: "banner_a"},
	)
	got := buildShareURL("https://ref.example.com/r/ABC123", domain.ChannelZalo, utm)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	want := map[string]string{
		"channel":      "zalo",
		"utm_source":   "zalo",
		"utm_medium":   "referral",
		"utm_campaign": "sale_tet_2026",
		"utm_content":  "banner_a",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Errorf("%s = %q, want %q (url %s)", k, q.Get(k), v, got)
		}
	}
	if q.Has("utm_term") {
		t.Errorf("utm_term should be omitted: %s", got)
	}
	if u.Path != "/r/ABC123" {
		t.Errorf("path = %s", u.Path)
	}
}

func TestUTMParamsBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var body generateLinkRequest
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"owner_user_id":"u1","utm_source":"Google","utm_medium":"cpc","utm_campaign":"x"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	if err := c.ShouldBindJSON(&body); err != nil {
		t.Fatal(err)
	}
	if got := body.toDomain(); got.Source != "google" || got.Medium != "cpc" || got.Campaign != "x" {
		t.Fatalf("json binding: %+v", got)
	}

	var q utmParams
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/r/ABC?utm_source=newsletter&utm_term=shoes", nil)
	if err := c.ShouldBindQuery(&q); err != nil {
		t.Fatal(err)
	}
	if got := q.toDomain(); got.Source != "newsletter" || got.Term != "shoes" {
		t.Fatalf("query binding: %+v", got)
	}
}

func TestBuildBranchURL(t *testing.T) {
	cfg := appconfig.DeepLinkConfig{
		PublicURL:   "https://ref.example.com",
		AppStoreURL: "https://apps.apple.com/app/id1",
		Branch:      appconfig.BranchConfig{Key: "key_live_abc", LinkDomain: "example.app.link"},
	}
	utm := domain.UTM{Source: "zalo", Medium: "referral", Campaign: "sale_tet_2026", Content: "banner_a"}
	got := buildBranchURL(cfg, "ABC12345", "click-1", domain.ChannelZalo, utm)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "example.app.link" || u.Path != "/a/key_live_abc" {
		t.Fatalf("unexpected branch url %s", got)
	}
	q := u.Query()
	want := map[string]string{
		"ref_code":     "ABC12345",
		"click_id":     "click-1",
		"utm_source":   "zalo",
		"utm_medium":   "referral",
		"utm_campaign": "sale_tet_2026",
		"utm_content":  "banner_a",
		"~channel":     "zalo",
		"~campaign":    "sale_tet_2026",
		"~feature":     "referral",
		"$ios_url":     "https://apps.apple.com/app/id1",
		"$desktop_url": "https://ref.example.com?utm_campaign=sale_tet_2026&utm_content=banner_a&utm_medium=referral&utm_source=zalo",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Errorf("%s = %q, want %q", k, q.Get(k), v)
		}
	}
	if q.Has("$android_url") {
		t.Error("$android_url should be omitted when PLAY_STORE_URL is empty")
	}
}
