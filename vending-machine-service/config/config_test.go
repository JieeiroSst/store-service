package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Postgres.Host != "postgresdb" || cfg.Vending.ReconcileAfter != 2*time.Minute ||
		cfg.Upstream.WalletServiceURL != "http://payment-wallet-service-svc" {
		t.Fatalf("config %+v", cfg)
	}
}

func TestLoadMergesConsulThenEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/kv/vending_test" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		rw.Write([]byte(`{"postgres":{"host":"db-from-consul","password":"from-consul"},"vending":{"sessionTtl":"7m"}}`))
	}))
	defer srv.Close()

	t.Setenv("CONSUL_ADDR", srv.URL)
	t.Setenv("CONSUL_KEY", "vending_test")
	t.Setenv("POSTGRES_PASSWORD", "from-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Postgres.Host != "db-from-consul" || cfg.Postgres.Password != "from-secret" {
		t.Fatalf("postgres %+v", cfg.Postgres)
	}
	if cfg.Vending.SessionTTL != 7*time.Minute || cfg.Vending.ReservationTTL != 2*time.Minute {
		t.Fatalf("vending %+v", cfg.Vending)
	}
}

func TestLoadWithoutConsulKeyUsesEnv(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	t.Setenv("CONSUL_ADDR", srv.URL)
	t.Setenv("POSTGRES_HOST", "env-db")
	cfg, err := Load()
	if err != nil || cfg.Postgres.Host != "env-db" {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
}

func TestValidateRejectsShortReconcileGrace(t *testing.T) {
	cfg := Defaults()
	cfg.Vending.ReconcileAfter = 30 * time.Second
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "reconcileAfter") {
		t.Fatalf("got %v", err)
	}
}

func TestMergeRejectsBadDuration(t *testing.T) {
	if err := Defaults().Merge([]byte(`{"vending":{"sessionTtl":300}}`)); err == nil {
		t.Fatal("expected error")
	}
}
