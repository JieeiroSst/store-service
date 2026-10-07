package config

import (
	"os"
	"testing"
	"time"
)

const testKey = "config-test-document-hash-key-0123456789"

func TestConsulJSONMatchesConfig(t *testing.T) {
	raw, err := os.ReadFile("../consul.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Defaults()
	if err := cfg.Merge(raw); err != nil {
		t.Fatal(err)
	}
	cfg.Security.DocumentHashKey = testKey
	cfg.Postgres.User = "customer_info_svc"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Storage.Backend != "postgres" || cfg.UserService.BaseURL != "http://user-api-svc" || cfg.UserService.Timeout != 5*time.Second ||
		cfg.EkycService.BaseURL != "http://ekyc-service-svc" || cfg.EkycService.Timeout != 60*time.Second ||
		!cfg.KYC.RequireFaceMatch || cfg.KYC.MinAge != 18 || cfg.Postgres.Database != "customer_info" {
		t.Fatalf("config %+v", cfg)
	}
}

func TestEnvOverridesAndValidate(t *testing.T) {
	t.Setenv("CONSUL_ADDR", "")
	t.Setenv("DOCUMENT_HASH_KEY", testKey)
	t.Setenv("USER_SERVICE_URL", "http://users:1235/")
	t.Setenv("USER_SERVICE_TOKEN", "t")
	t.Setenv("EKYC_SERVICE_URL", "http://ekyc:8087/")
	t.Setenv("KYC_REQUIRE_FACE_MATCH", "false")
	t.Setenv("KYC_MIN_AGE", "21")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UserService.BaseURL != "http://users:1235" || cfg.UserService.Token != "t" || cfg.EkycService.BaseURL != "http://ekyc:8087" || cfg.KYC.RequireFaceMatch || cfg.KYC.MinAge != 21 {
		t.Fatalf("config %+v", cfg)
	}
	bad := map[string]func(*Config){
		"short key":    func(c *Config) { c.Security.DocumentHashKey = "x" },
		"relative url": func(c *Config) { c.UserService.BaseURL = "users" },
		"confidence":   func(c *Config) { c.KYC.MinConfidence = 2 },
		"pg no user":   func(c *Config) { c.Storage.Backend = "postgres" },
		"bad backend":  func(c *Config) { c.Storage.Backend = "redis" },
	}
	for name, mut := range bad {
		c := Defaults()
		c.Security.DocumentHashKey = testKey
		mut(c)
		if c.Validate() == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
