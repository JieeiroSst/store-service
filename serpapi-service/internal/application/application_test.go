package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/adapter/secondary/cache"
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

type fakeAPI struct {
	mu       sync.Mutex
	searches []domain.SearchRequest
	status   string
	err      error
}

func (f *fakeAPI) Search(_ context.Context, req domain.SearchRequest) (domain.SearchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searches = append(f.searches, req)
	if f.err != nil {
		return domain.SearchResult{}, f.err
	}
	return domain.SearchResult{Body: []byte(`{"q":"` + req.Params["q"] + `"}`), ContentType: "application/json", SearchID: "id1", Status: f.status}, nil
}

func (f *fakeAPI) Archive(_ context.Context, id string, _ domain.Output) (domain.SearchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searches = append(f.searches, domain.SearchRequest{Engine: "archive:" + id})
	return domain.SearchResult{Body: []byte(`{}`), SearchID: id, Status: f.status}, nil
}

func (f *fakeAPI) Account(context.Context) (domain.Account, error) { return domain.Account{}, nil }

func (f *fakeAPI) Locations(context.Context, domain.LocationQuery) ([]domain.Location, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searches = append(f.searches, domain.SearchRequest{Engine: "locations"})
	return []domain.Location{{ID: "1"}}, nil
}

func (f *fakeAPI) UploadImage(_ context.Context, img domain.ImageUpload) (domain.UploadedImage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searches = append(f.searches, domain.SearchRequest{Engine: "image:" + img.ContentType})
	return domain.UploadedImage{ImageID: "img1"}, nil
}

func (f *fakeAPI) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.searches)
}

type nopMetrics struct{}

func (nopMetrics) ObserveUpstream(string, string, string, time.Duration) {}
func (nopMetrics) ObserveCache(string, bool)                             {}

func newServices(strict bool) (*SearchService, *AccountService, *fakeAPI) {
	cfg := &config.Config{
		Search: config.SearchConfig{StrictEngines: strict, MaxConcurrent: 4, MaxBatchSize: 3},
		Cache:  config.CacheConfig{TTL: time.Hour, LocationsTTL: time.Hour},
	}
	api := &fakeAPI{status: domain.StatusSuccess}
	c := cache.NewMemory(100)
	lim := NewLimiter(cfg)
	return NewSearchService(cfg, domain.NewCatalog(domain.Engines), api, c, nopMetrics{}, lim),
		NewAccountService(cfg, api, c, nopMetrics{}, lim), api
}

func TestSearchCachesSuccessfulResults(t *testing.T) {
	s, _, api := newServices(false)
	ctx := context.Background()
	req := domain.SearchRequest{Engine: "google", Params: domain.Params{"q": "go"}}
	first, err := s.Search(ctx, req)
	if err != nil || first.Cached {
		t.Fatalf("first: %+v %v", first, err)
	}
	second, err := s.Search(ctx, req)
	if err != nil || !second.Cached || string(second.Body) != string(first.Body) || second.SearchID != "id1" {
		t.Fatalf("second: %+v %v", second, err)
	}
	if api.calls() != 1 {
		t.Fatalf("upstream calls = %d", api.calls())
	}
	req.NoCache = true
	if res, _ := s.Search(ctx, req); res.Cached || api.calls() != 2 {
		t.Fatal("no_cache must bypass the cache")
	}
}

func TestSearchSkipsCacheForAsyncZeroTraceAndFailures(t *testing.T) {
	s, _, api := newServices(false)
	ctx := context.Background()
	for _, req := range []domain.SearchRequest{
		{Engine: "google", Params: domain.Params{"q": "a"}, Async: true},
		{Engine: "google", Params: domain.Params{"q": "b", "zero_trace": "true"}},
	} {
		_, _ = s.Search(ctx, req)
		if res, _ := s.Search(ctx, req); res.Cached {
			t.Errorf("%+v was cached", req)
		}
	}
	api.status = "Error"
	req := domain.SearchRequest{Engine: "google", Params: domain.Params{"q": "c"}}
	_, _ = s.Search(ctx, req)
	if res, _ := s.Search(ctx, req); res.Cached {
		t.Error("error status was cached")
	}
}

func TestSearchValidation(t *testing.T) {
	s, _, api := newServices(false)
	ctx := context.Background()
	if _, err := s.Search(ctx, domain.SearchRequest{Engine: "google"}); !domain.IsInvalid(err) {
		t.Errorf("missing q: %v", err)
	}
	if _, err := s.Search(ctx, domain.SearchRequest{Engine: "Google; drop"}); !domain.IsInvalid(err) {
		t.Errorf("bad engine id: %v", err)
	}
	if _, err := s.Search(ctx, domain.SearchRequest{Engine: "brand_new_engine"}); err != nil {
		t.Errorf("unknown engine should pass through: %v", err)
	}
	_, _ = s.Search(ctx, domain.SearchRequest{Engine: "bing", Params: domain.Params{"q": "x", "api_key": "stolen", "engine": "google"}})
	last := api.searches[len(api.searches)-1]
	if _, ok := last.Params["api_key"]; ok || last.Engine != "bing" || last.Output != domain.OutputJSON {
		t.Errorf("reserved params not stripped: %+v", last)
	}

	strict, _, _ := newServices(true)
	if _, err := strict.Search(ctx, domain.SearchRequest{Engine: "brand_new_engine"}); !domain.IsInvalid(err) {
		t.Errorf("strict mode accepted unknown engine: %v", err)
	}
}

func TestBatch(t *testing.T) {
	s, _, _ := newServices(false)
	ctx := context.Background()
	items, err := s.Batch(ctx, []domain.SearchRequest{
		{Engine: "google", Params: domain.Params{"q": "a"}},
		{Engine: "google"},
	})
	if err != nil || len(items) != 2 || items[0].Err != nil || !domain.IsInvalid(items[1].Err) {
		t.Fatalf("batch: %+v %v", items, err)
	}
	if _, err := s.Batch(ctx, make([]domain.SearchRequest, 4)); !domain.IsInvalid(err) {
		t.Fatalf("oversized batch: %v", err)
	}
}

func TestArchiveCachesOnlyFinishedSearches(t *testing.T) {
	s, _, api := newServices(false)
	ctx := context.Background()
	api.status = "Processing"
	_, _ = s.Archive(ctx, "abc", domain.OutputJSON)
	_, _ = s.Archive(ctx, "abc", domain.OutputJSON)
	if api.calls() != 2 {
		t.Fatalf("processing search was cached")
	}
	api.status = domain.StatusSuccess
	_, _ = s.Archive(ctx, "abc", domain.OutputJSON)
	if res, _ := s.Archive(ctx, "abc", domain.OutputJSON); !res.Cached {
		t.Fatal("finished search not cached")
	}
	if _, err := s.Archive(ctx, "../account", domain.OutputJSON); !domain.IsInvalid(err) {
		t.Fatalf("path traversal: %v", err)
	}
	if _, err := s.Archive(ctx, "abc", domain.OutputMarkdown); !domain.IsInvalid(err) {
		t.Fatalf("md archive: %v", err)
	}
}

func TestLocationsCached(t *testing.T) {
	_, a, api := newServices(false)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if locs, err := a.Locations(ctx, domain.LocationQuery{Q: "Austin"}); err != nil || len(locs) != 1 {
			t.Fatal(locs, err)
		}
	}
	if api.calls() != 1 {
		t.Fatalf("calls = %d", api.calls())
	}
	if _, err := a.Locations(ctx, domain.LocationQuery{Limit: 1000}); !domain.IsInvalid(err) {
		t.Fatal(err)
	}
}

func TestLimiterHonorsContext(t *testing.T) {
	l := NewLimiter(&config.Config{Search: config.SearchConfig{MaxConcurrent: 1}})
	release, _ := l.Acquire(context.Background())
	defer release()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := l.Acquire(ctx); err == nil {
		t.Fatal("acquired a full limiter")
	}
}

func TestSearchOptionRules(t *testing.T) {
	s, _, _ := newServices(false)
	ctx := context.Background()
	if _, err := s.Search(ctx, domain.SearchRequest{Engine: "google", Params: domain.Params{"q": "x"}, Async: true, NoCache: true}); !domain.IsInvalid(err) {
		t.Errorf("async+no_cache: %v", err)
	}
	if _, err := s.Search(ctx, domain.SearchRequest{Engine: "bing", Params: domain.Params{"q": "x"}, Output: domain.OutputPixelPosition}); !domain.IsInvalid(err) {
		t.Errorf("pixel position on bing: %v", err)
	}
	req := domain.SearchRequest{Engine: "google", Params: domain.Params{"q": "x"}, Output: domain.OutputPixelPosition}
	_, _ = s.Search(ctx, req)
	if res, err := s.Search(ctx, req); err != nil || !res.Cached {
		t.Errorf("pixel position result not cached: %+v %v", res, err)
	}
	if _, err := s.Archive(ctx, "abc", domain.OutputPixelPosition); err != nil {
		t.Errorf("pixel position archive: %v", err)
	}
}

func TestImageUpload(t *testing.T) {
	_, _, api := newServices(false)
	cfg := &config.Config{Search: config.SearchConfig{MaxConcurrent: 1}}
	svc := NewImageService(api, nopMetrics{}, NewLimiter(cfg))
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	ctx := context.Background()

	out, err := svc.Upload(ctx, "", []byte("\x89PNG\r\n\x1a\nrest"))
	if err != nil || out.ImageID != "img1" || !out.ExpiresAt.Equal(now.Add(domain.ImageIDTTL)) {
		t.Fatalf("upload: %+v %v", out, err)
	}
	if api.searches[len(api.searches)-1].Engine != "image:image/png" {
		t.Fatalf("content type not sniffed: %+v", api.searches)
	}
	if _, err := svc.Upload(ctx, "a.gif", []byte("GIF89a")); !domain.IsInvalid(err) {
		t.Errorf("gif accepted: %v", err)
	}
	big := append([]byte{0xFF, 0xD8, 0xFF}, make([]byte, domain.MaxImageBytes)...)
	if _, err := svc.Upload(ctx, "a.jpg", big); !domain.IsInvalid(err) {
		t.Errorf("oversized image accepted: %v", err)
	}
	if api.calls() != 1 {
		t.Errorf("invalid uploads reached SerpApi: %d calls", api.calls())
	}
}
