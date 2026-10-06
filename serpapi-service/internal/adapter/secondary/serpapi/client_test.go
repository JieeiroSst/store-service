package serpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

const key = "secret-key"

func newClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cfg := &config.Config{SerpAPI: config.SerpAPIConfig{BaseURL: srv.URL, APIKey: key, Timeout: 5 * time.Second, MaxRetries: 2}}
	c := NewClient(cfg)
	c.backoff = time.Millisecond
	return c
}

func TestSearchSendsParamsAndParsesMetadata(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/search" || q.Get("api_key") != key || q.Get("engine") != "google" ||
			q.Get("q") != "coffee" || q.Get("output") != "json" || q.Get("no_cache") != "true" {
			t.Errorf("unexpected request %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"search_metadata":{"id":"abc","status":"Success"},"organic_results":[]}`))
	})
	res, err := c.Search(context.Background(), domain.SearchRequest{
		Engine: "google", Params: domain.Params{"q": "coffee"}, Output: domain.OutputJSON, NoCache: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.SearchID != "abc" || res.Status != "Success" || !strings.Contains(string(res.Body), "organic_results") {
		t.Fatalf("bad result %+v", res)
	}
}

func TestArchiveAccountLocations(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/searches/abc.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html></html>"))
		case "/account.json":
			_, _ = w.Write([]byte(`{"account_id":"1","api_key":"` + key + `","plan_searches_left":42}`))
		case "/locations.json":
			if r.URL.Query().Has("api_key") {
				t.Error("locations API must not receive the API key")
			}
			if r.URL.Query().Get("q") != "Austin" || r.URL.Query().Get("limit") != "3" {
				t.Errorf("bad locations query %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[{"id":"585069bdee19ad271e9bc072","canonical_name":"Austin,TX,Texas,United States","gps":[-97.7,30.2]}]`))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()
	res, err := c.Archive(ctx, "abc", domain.OutputHTML)
	if err != nil || res.ContentType != "text/html" {
		t.Fatalf("archive: %+v %v", res, err)
	}
	acc, err := c.Account(ctx)
	if err != nil || acc.PlanSearchesLeft != 42 {
		t.Fatalf("account: %+v %v", acc, err)
	}
	locs, err := c.Locations(ctx, domain.LocationQuery{Q: "Austin", Limit: 3})
	if err != nil || len(locs) != 1 || locs[0].CanonicalName == "" {
		t.Fatalf("locations: %+v %v", locs, err)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		check  func(error) bool
	}{
		{400, domain.IsInvalid},
		{404, func(err error) bool { return errors.Is(err, domain.ErrNotFound) }},
		{410, func(err error) bool { return errors.Is(err, domain.ErrExpired) }},
		{429, func(err error) bool { return errors.Is(err, domain.ErrRateLimited) }},
		{401, domain.IsUpstream},
		{502, domain.IsUpstream},
	}
	for _, tc := range cases {
		c := newClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
		})
		_, err := c.Search(context.Background(), domain.SearchRequest{Engine: "google", Output: domain.OutputJSON})
		if err == nil || !tc.check(err) || !strings.Contains(err.Error(), "boom") {
			t.Errorf("status %d: got %v", tc.status, err)
		}
	}
}

func TestRetriesServerErrorsOnly(t *testing.T) {
	var calls atomic.Int32
	c := newClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"search_metadata":{"id":"x","status":"Success"}}`))
	})
	if _, err := c.Search(context.Background(), domain.SearchRequest{Engine: "google", Output: domain.OutputJSON}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", calls.Load())
	}

	calls.Store(0)
	c = newClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	})
	_, _ = c.Search(context.Background(), domain.SearchRequest{Engine: "google", Output: domain.OutputJSON})
	if calls.Load() != 1 {
		t.Fatalf("4xx retried: calls = %d", calls.Load())
	}
}

func TestNetworkErrorDoesNotLeakKey(t *testing.T) {
	cfg := &config.Config{SerpAPI: config.SerpAPIConfig{BaseURL: "http://127.0.0.1:1", APIKey: key, Timeout: time.Second}}
	_, err := NewClient(cfg).Account(context.Background())
	if err == nil || !domain.IsUpstream(err) {
		t.Fatalf("want upstream error, got %v", err)
	}
	if strings.Contains(err.Error(), key) {
		t.Fatalf("error leaks the API key: %v", err)
	}
}

func TestUploadImageAndPixelArchive(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/image":
			if r.URL.RawQuery != "" {
				t.Errorf("api key must be a form field, got query %q", r.URL.RawQuery)
			}
			if r.FormValue("api_key") != key {
				t.Errorf("api_key form field = %q", r.FormValue("api_key"))
			}
			f, h, err := r.FormFile("image")
			if err != nil || h.Filename != "a.png" || h.Header.Get("Content-Type") != "image/png" {
				t.Errorf("image part: %v %+v", err, h)
			} else {
				f.Close()
			}
			_, _ = w.Write([]byte(`{"message":"Image uploaded successfully.","image_id":"img42"}`))
		case r.URL.Path == "/searches/abc.json_with_pixel_position":
			_, _ = w.Write([]byte(`{"search_metadata":{"id":"abc","status":"Success"},"pixel_position_information":{}}`))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()
	out, err := c.UploadImage(ctx, domain.ImageUpload{Filename: "a.png", ContentType: "image/png", Data: []byte("png")})
	if err != nil || out.ImageID != "img42" {
		t.Fatalf("upload: %+v %v", out, err)
	}
	res, err := c.Archive(ctx, "abc", domain.OutputPixelPosition)
	if err != nil || res.SearchID != "abc" || res.Status != "Success" {
		t.Fatalf("pixel archive: %+v %v", res, err)
	}

	c = newClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":"Invalid image format. Supported format: jpg, jpeg, png, webp"}`))
	})
	if _, err := c.UploadImage(ctx, domain.ImageUpload{Filename: "a.png", ContentType: "image/png", Data: []byte("x")}); !domain.IsInvalid(err) {
		t.Fatalf("upload error body: %v", err)
	}
}
