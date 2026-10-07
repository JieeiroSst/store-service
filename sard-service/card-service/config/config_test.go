package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const testKey = "config-test-master-key-0123456789abcdef"

func TestConsulJSONMatchesConfig(t *testing.T) {
	raw, err := os.ReadFile("../consul.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Defaults()
	if err := cfg.Merge(raw); err != nil {
		t.Fatal(err)
	}
	cfg.Security.MasterKey = testKey
	cfg.Postgres.User = "card_svc"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Storage.Backend != "postgres" || cfg.Postgres.Host != "postgresdb" || cfg.Postgres.Port != "80" ||
		cfg.Postgres.Database != "card" || cfg.Postgres.MaxConns != 10 || cfg.Limits.Location.String() != "Asia/Ho_Chi_Minh" ||
		cfg.CustomerInfo.BaseURL != "http://sard-customer-info-service-svc:8080" || cfg.AuthService.Timeout != 5*time.Second {
		t.Fatalf("config %+v", cfg)
	}
}

func TestLoadMergesConsulThenEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/kv/card_test" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		rw.Write([]byte(`{"storage":{"backend":"postgres"},"postgres":{"host":"db","port":"5432","user":"from-consul","password":"p@ss/word","database":"card"}}`))
	}))
	defer srv.Close()

	t.Setenv("CONSUL_ADDR", srv.URL)
	t.Setenv("CONSUL_KEY", "card_test")
	t.Setenv("CARD_MASTER_KEY", testKey)
	t.Setenv("POSTGRES_USER", "from-env")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("ACCESS_TOKENS", " a, ,b ")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Postgres.User != "from-env" || cfg.Postgres.Password != "p@ss/word" || len(cfg.Server.AccessTokens) != 2 {
		t.Fatalf("config %+v", cfg)
	}
	dsn := cfg.Postgres.DSN()
	if !strings.HasPrefix(dsn, "postgres://from-env:p%40ss%2Fword@db:5432/card?") || !strings.Contains(dsn, "sslmode=disable") {
		t.Fatalf("dsn %s", dsn)
	}
}

func TestValidate(t *testing.T) {
	cases := map[string]func(*Config){
		"short master key": func(c *Config) { c.Security.MasterKey = "short" },
		"unknown backend":  func(c *Config) { c.Storage.Backend = "mysql" },
		"postgres no user": func(c *Config) { c.Storage.Backend = "postgres" },
		"bad timezone":     func(c *Config) { c.Limits.Timezone = "Mars/Olympus" },
		"relative url":     func(c *Config) { c.AuthService.BaseURL = "auth" },
	}
	for name, mut := range cases {
		cfg := Defaults()
		cfg.Security.MasterKey = testKey
		mut(cfg)
		if cfg.Validate() == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	cfg := Defaults()
	cfg.Security.MasterKey = testKey
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}
