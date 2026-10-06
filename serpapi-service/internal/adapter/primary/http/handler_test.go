package http

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/adapter/secondary/cache"
	"github.com/JIeeiroSst/serpapi-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/serpapi-service/internal/adapter/secondary/serpapi"
	"github.com/JIeeiroSst/serpapi-service/internal/application"
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

const apiKey = "upstream-secret"

func newAPI(t *testing.T, tokens ...string) *httptest.Server {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("api_key") != apiKey && r.FormValue("api_key") != apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Invalid API key."}`))
			return
		}
		switch r.URL.Path {
		case "/search":
			if q.Get("output") == "md" {
				w.Header().Set("Content-Type", "text/markdown")
				_, _ = w.Write([]byte("# " + q.Get("q")))
				return
			}
			if q.Get("output") == "html" {
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte("<html>" + q.Get("q") + "</html>"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"search_metadata":   map[string]string{"id": "sid-" + q.Get("engine"), "status": "Success"},
				"search_parameters": map[string]string{"engine": q.Get("engine"), "q": q.Get("q"), "num": q.Get("num")},
			})
		case "/image":
			if f, _, err := r.FormFile("image"); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				f.Close()
			}
			_, _ = w.Write([]byte(`{"message":"Image uploaded successfully.","image_id":"img1"}`))
		case "/searches/old.json":
			w.WriteHeader(http.StatusGone)
			_, _ = w.Write([]byte(`{"error":"The search expired and has been deleted from the archive."}`))
		case "/account.json":
			_, _ = w.Write([]byte(`{"account_id":"acc","api_key":"` + apiKey + `","total_searches_left":7}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Search not found."}`))
		}
	}))
	t.Cleanup(upstream.Close)

	cfg := &config.Config{
		Server:  config.ServerConfig{AccessTokens: tokens},
		SerpAPI: config.SerpAPIConfig{BaseURL: upstream.URL, APIKey: apiKey, Timeout: 5 * time.Second},
		Search:  config.SearchConfig{MaxConcurrent: 4, MaxBatchSize: 5},
		Cache:   config.CacheConfig{TTL: time.Hour, LocationsTTL: time.Hour},
	}
	client := serpapi.NewClient(cfg)
	c := cache.NewMemory(100)
	m := metrics.NewPrometheus()
	lim := application.NewLimiter(cfg)
	h := NewHandler(cfg,
		application.NewSearchService(cfg, domain.NewCatalog(domain.Engines), client, c, m, lim),
		application.NewAccountService(cfg, client, c, m, lim),
		application.NewImageService(client, m, lim))
	srv := httptest.NewServer(NewRouter(h))
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, method, url, body string, header ...string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func TestSearchEndpoints(t *testing.T) {
	srv := newAPI(t)

	resp, body := do(t, "GET", srv.URL+"/api/v1/search?engine=google&q=coffee&api_key=attacker", "")
	if resp.StatusCode != 200 || resp.Header.Get("X-Cache") != "MISS" || resp.Header.Get("X-Search-Id") != "sid-google" {
		t.Fatalf("GET search: %d %v %s", resp.StatusCode, resp.Header, body)
	}
	resp, _ = do(t, "GET", srv.URL+"/api/v1/search/google?q=coffee", "")
	if resp.Header.Get("X-Cache") != "HIT" {
		t.Fatal("path form of the same search should hit the cache")
	}

	resp, body = do(t, "POST", srv.URL+"/api/v1/search", `{"engine":"bing","params":{"q":"tea","num":20}}`)
	if resp.StatusCode != 200 || !strings.Contains(body, `"num":"20"`) {
		t.Fatalf("POST search: %d %s", resp.StatusCode, body)
	}

	resp, body = do(t, "GET", srv.URL+"/api/v1/search?engine=google&q=html&output=html", "")
	if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "text/html" || body != "<html>html</html>" {
		t.Fatalf("html: %d %s", resp.StatusCode, body)
	}

	resp, body = do(t, "GET", srv.URL+"/api/v1/search?engine=google&q=md", "", "Accept", "text/markdown")
	if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "text/markdown" || body != "# md" {
		t.Fatalf("markdown via Accept: %d %v %s", resp.StatusCode, resp.Header, body)
	}

	resp, body = do(t, "GET", srv.URL+"/api/v1/search?engine=google", "")
	if resp.StatusCode != 400 || !strings.Contains(body, "requires parameter(s): q") {
		t.Fatalf("missing q: %d %s", resp.StatusCode, body)
	}

	resp, body = do(t, "POST", srv.URL+"/api/v1/search/batch",
		`{"searches":[{"engine":"google","params":{"q":"a"}},{"engine":"youtube","params":{}}]}`)
	var batch batchResponseDTO
	if err := json.Unmarshal([]byte(body), &batch); err != nil || resp.StatusCode != 200 {
		t.Fatalf("batch: %d %s", resp.StatusCode, body)
	}
	if len(batch.Results) != 2 || batch.Results[0].Status != 200 || batch.Results[1].Status != 400 {
		t.Fatalf("batch results: %+v", batch.Results)
	}

	resp, _ = do(t, "GET", srv.URL+"/api/v1/searches/unknown", "")
	if resp.StatusCode != 404 {
		t.Fatalf("archive miss: %d", resp.StatusCode)
	}
	resp, _ = do(t, "GET", srv.URL+"/api/v1/searches/old", "")
	if resp.StatusCode != 410 {
		t.Fatalf("expired archive: %d", resp.StatusCode)
	}
	resp, body = do(t, "GET", srv.URL+"/api/v1/search?engine=bing&q=x&output=json_with_pixel_position", "")
	if resp.StatusCode != 400 {
		t.Fatalf("pixel position on bing: %d %s", resp.StatusCode, body)
	}
}

func TestEnginesAndAccount(t *testing.T) {
	srv := newAPI(t)
	resp, body := do(t, "GET", srv.URL+"/api/v1/engines?group=Walmart", "")
	if resp.StatusCode != 200 || !strings.Contains(body, `"count":4`) {
		t.Fatalf("engines: %s", body)
	}
	resp, _ = do(t, "GET", srv.URL+"/api/v1/engines/nope", "")
	if resp.StatusCode != 404 {
		t.Fatalf("unknown engine: %d", resp.StatusCode)
	}
	resp, body = do(t, "GET", srv.URL+"/api/v1/account", "")
	if resp.StatusCode != 200 || strings.Contains(body, apiKey) || !strings.Contains(body, `"total_searches_left":7`) {
		t.Fatalf("account: %d %s", resp.StatusCode, body)
	}
}

func TestAccessTokens(t *testing.T) {
	srv := newAPI(t, "t1", "t2")
	if resp, _ := do(t, "GET", srv.URL+"/api/v1/engines", ""); resp.StatusCode != 401 {
		t.Fatalf("no token: %d", resp.StatusCode)
	}
	if resp, _ := do(t, "GET", srv.URL+"/api/v1/engines", "", "Authorization", "Bearer bad"); resp.StatusCode != 401 {
		t.Fatalf("bad token: %d", resp.StatusCode)
	}
	if resp, _ := do(t, "GET", srv.URL+"/api/v1/engines", "", "Authorization", "Bearer t2"); resp.StatusCode != 200 {
		t.Fatalf("good token: %d", resp.StatusCode)
	}
	if resp, _ := do(t, "GET", srv.URL+"/api/v1/engines", "", "X-Api-Token", "t1"); resp.StatusCode != 200 {
		t.Fatalf("header token: %d", resp.StatusCode)
	}
	if resp, _ := do(t, "GET", srv.URL+"/health", ""); resp.StatusCode != 200 {
		t.Fatalf("health must stay open: %d", resp.StatusCode)
	}
}

func TestUploadImage(t *testing.T) {
	srv := newAPI(t)
	png := "\x89PNG\r\n\x1a\nimagedata"

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("image", "../../etc/cat.png")
	_, _ = fw.Write([]byte(png))
	_ = mw.Close()
	resp, body := do(t, "POST", srv.URL+"/api/v1/images", buf.String(), "Content-Type", mw.FormDataContentType())
	if resp.StatusCode != 201 || !strings.Contains(body, `"image_id":"img1"`) || !strings.Contains(body, "expires_at") {
		t.Fatalf("multipart upload: %d %s", resp.StatusCode, body)
	}

	resp, body = do(t, "POST", srv.URL+"/api/v1/images", png, "Content-Type", "image/png")
	if resp.StatusCode != 201 {
		t.Fatalf("raw upload: %d %s", resp.StatusCode, body)
	}

	resp, body = do(t, "POST", srv.URL+"/api/v1/images", "GIF89a", "Content-Type", "image/gif")
	if resp.StatusCode != 400 {
		t.Fatalf("gif: %d %s", resp.StatusCode, body)
	}

	big := "\xFF\xD8\xFF" + strings.Repeat("a", domain.MaxImageBytes+maxMultipartOverhead)
	resp, body = do(t, "POST", srv.URL+"/api/v1/images", big, "Content-Type", "image/jpeg")
	if resp.StatusCode != 400 || !strings.Contains(body, "larger than") {
		t.Fatalf("oversized: %d %s", resp.StatusCode, body)
	}

	buf.Reset()
	mw = multipart.NewWriter(&buf)
	_ = mw.WriteField("other", "x")
	_ = mw.Close()
	resp, body = do(t, "POST", srv.URL+"/api/v1/images", buf.String(), "Content-Type", mw.FormDataContentType())
	if resp.StatusCode != 400 || !strings.Contains(body, `\"image\" is required`) {
		t.Fatalf("missing part: %d %s", resp.StatusCode, body)
	}
}
