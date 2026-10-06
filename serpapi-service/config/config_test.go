package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestConsulJSONMatchesConfig(t *testing.T) {
	raw, err := os.ReadFile("../consul.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Defaults()
	if err := cfg.Merge(raw); err != nil {
		t.Fatal(err)
	}
	cfg.SerpAPI.APIKey = "k"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.SerpAPI.Timeout != 90*time.Second || cfg.Cache.LocationsTTL != 24*time.Hour ||
		cfg.Cache.RedisAddr != "redis-svc:6379" || cfg.Search.MaxBatchSize != 20 {
		t.Fatalf("config %+v", cfg)
	}
}

func TestLoadMergesConsulThenEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/kv/serpapi_test" || r.Header.Get("X-Consul-Token") != "ct" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		rw.Write([]byte(`{"serpapi":{"baseUrl":"https://consul.example/","timeout":"30s","apiKey":"from-consul"},
			"search":{"strictEngines":true},"cache":{"backend":"redis","ttl":"10m"}}`))
	}))
	defer srv.Close()

	t.Setenv("CONSUL_ADDR", srv.URL)
	t.Setenv("CONSUL_KEY", "serpapi_test")
	t.Setenv("CONSUL_HTTP_TOKEN", "ct")
	t.Setenv("SERPAPI_API_KEY", "from-secret")
	t.Setenv("ACCESS_TOKENS", "a, b")
	t.Setenv("CACHE_TTL", "20m")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SerpAPI.BaseURL != "https://consul.example" || cfg.SerpAPI.Timeout != 30*time.Second {
		t.Fatalf("serpapi %+v", cfg.SerpAPI)
	}

	if cfg.SerpAPI.APIKey != "from-secret" || len(cfg.Server.AccessTokens) != 2 {
		t.Fatalf("secrets %+v %v", cfg.SerpAPI, cfg.Server.AccessTokens)
	}
	if !cfg.Search.StrictEngines || cfg.Cache.Backend != "redis" || cfg.Cache.TTL != 20*time.Minute ||
		cfg.Cache.LocationsTTL != 24*time.Hour {
		t.Fatalf("search/cache %+v %+v", cfg.Search, cfg.Cache)
	}
}

func TestLoadWithoutConsulKeyUsesDefaultsAndEnv(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	t.Setenv("CONSUL_ADDR", srv.URL)
	t.Setenv("SERPAPI_API_KEY", "k")
	t.Setenv("PORT_HTTP_SERVER", "9000")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.PortHttpServer != "9000" || cfg.Cache.Backend != "memory" {
		t.Fatalf("config %+v", cfg)
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	t.Setenv("CONSUL_ADDR", "")
	t.Setenv("SERPAPI_API_KEY", "")
	if _, err := Load(); err == nil {
		t.Fatal("missing API key accepted")
	}
	t.Setenv("SERPAPI_API_KEY", "k")
	t.Setenv("CACHE_BACKEND", "disk")
	if _, err := Load(); err == nil {
		t.Fatal("unknown cache backend accepted")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write([]byte(`{"serpapi":{"timeout":90}}`))
	}))
	defer srv.Close()
	t.Setenv("CACHE_BACKEND", "memory")
	t.Setenv("CONSUL_ADDR", srv.URL)
	if _, err := Load(); err == nil {
		t.Fatal("numeric duration accepted")
	}
}
