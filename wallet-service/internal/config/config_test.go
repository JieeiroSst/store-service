package config

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeConsul(t *testing.T, key, value string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/v1/catalog/service/"):
			fmt.Fprint(w, "[]")
		case r.URL.Path == "/v1/kv/"+key:
			fmt.Fprintf(w, `[{"Key":%q,"Value":%q}]`, key, base64.StdEncoding.EncodeToString([]byte(value)))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func inTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	return dir
}

func clearConsulEnv(t *testing.T) {
	for _, k := range []string{"HostConsul", "KeyConsul", "ServiceConsul"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

func TestLoadReadsConsulKeyFromEnvFile(t *testing.T) {
	clearConsulEnv(t)
	srv := fakeConsul(t, "wallet-service", `{"server":{"port":"9090"},"database":{"dsn":"postgres://consul"},"fees":{"withdrawal_rate":"0.01"}}`)
	dir := inTempDir(t)
	env := fmt.Sprintf("HostConsul=%q\nKeyConsul=\"wallet-service\"\nServiceConsul=\"consul\"\n", srv.URL)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REDIS_ADDR", "redis-from-env:6379")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "consul:wallet-service" {
		t.Fatalf("unexpected source %q", cfg.Source)
	}
	if cfg.Server.Port != ":9090" || cfg.Database.DSN != "postgres://consul" || cfg.Fees.WithdrawalRate != "0.01" {
		t.Fatalf("consul values not applied: %+v", cfg)
	}
	if cfg.Redis.Addr != "redis-from-env:6379" {
		t.Fatalf("env vars must override consul, got %q", cfg.Redis.Addr)
	}
	if cfg.Security.PinBcryptCost != 10 {
		t.Fatalf("defaults must fill missing keys, got %d", cfg.Security.PinBcryptCost)
	}
}

func TestLoadFailsWhenConsulKeyMissing(t *testing.T) {
	clearConsulEnv(t)
	srv := fakeConsul(t, "other", `{}`)
	inTempDir(t)
	t.Setenv("HostConsul", srv.URL)
	t.Setenv("KeyConsul", "wallet-service")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing key error, got %v", err)
	}
}

func TestLoadFallsBackToLocalFileWithoutConsul(t *testing.T) {
	clearConsulEnv(t)
	dir := inTempDir(t)
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte("database:\n  dsn: \"postgres://file\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cfg.Source, "file:") || cfg.Database.DSN != "postgres://file" {
		t.Fatalf("expected local file config, got %q %q", cfg.Source, cfg.Database.DSN)
	}
}

func TestConnStringFromFields(t *testing.T) {
	d := DatabaseConfig{Host: "postgresdb", Port: "80", User: "wallet_service_svc", Password: "p@ss:w/rd", Name: "wallet_service", SSLMode: "disable"}
	got, err := d.ConnString()
	if err != nil {
		t.Fatal(err)
	}
	want := "postgres://wallet_service_svc:p%40ss%3Aw%2Frd@postgresdb:80/wallet_service?sslmode=disable"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
	d.DSN = "postgres://explicit"
	if got, _ := d.ConnString(); got != "postgres://explicit" {
		t.Fatalf("dsn must win, got %s", got)
	}
	if _, err := (DatabaseConfig{}).ConnString(); err == nil {
		t.Fatalf("expected error without dsn or host")
	}
}

func TestEnvOverridesDatabaseCredentials(t *testing.T) {
	clearConsulEnv(t)
	srv := fakeConsul(t, "wallet-service", `{"database":{"host":"postgresdb","port":"80","user":"from-consul","password":"","name":"wallet_service"}}`)
	inTempDir(t)
	t.Setenv("HostConsul", srv.URL)
	t.Setenv("KeyConsul", "wallet-service")
	t.Setenv("DATABASE_USER", "wallet_service_svc")
	t.Setenv("DATABASE_PASSWORD", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.User != "wallet_service_svc" || cfg.Database.Password != "secret" || !cfg.Database.AutoMigrate {
		t.Fatalf("unexpected database config: %+v", cfg.Database)
	}
}
