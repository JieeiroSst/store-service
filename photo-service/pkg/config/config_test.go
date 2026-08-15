package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_ApplyDefaults_FillsZeroValues(t *testing.T) {
	var cfg Config
	cfg.applyDefaults()

	if cfg.App.Env != "development" {
		t.Errorf("App.Env = %q, want development", cfg.App.Env)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("HTTP.Port = %d, want 8080", cfg.HTTP.Port)
	}
	if time.Duration(cfg.HTTP.ShutdownTimeout) != 10*time.Second {
		t.Errorf("HTTP.ShutdownTimeout = %s, want 10s", time.Duration(cfg.HTTP.ShutdownTimeout))
	}
	if cfg.Postgres.MaxConns != 10 {
		t.Errorf("Postgres.MaxConns = %d, want 10", cfg.Postgres.MaxConns)
	}
	if cfg.Postgres.Host != "localhost" || cfg.Postgres.Port != 5432 {
		t.Errorf("Postgres.Host/Port = %s:%d, want localhost:5432", cfg.Postgres.Host, cfg.Postgres.Port)
	}
	if cfg.MinIO.Bucket != "photo-service" {
		t.Errorf("MinIO.Bucket = %q, want photo-service", cfg.MinIO.Bucket)
	}
	if time.Duration(cfg.MinIO.PresignTTL) != time.Hour {
		t.Errorf("MinIO.PresignTTL = %s, want 1h", time.Duration(cfg.MinIO.PresignTTL))
	}
}

func TestConfig_ApplyDefaults_PreservesExplicitValues(t *testing.T) {
	cfg := Config{MinIO: MinIOConfig{Bucket: "custom-bucket"}, HTTP: HTTPConfig{Port: 9090}}
	cfg.applyDefaults()

	if cfg.MinIO.Bucket != "custom-bucket" {
		t.Errorf("MinIO.Bucket = %q, want custom-bucket to be preserved", cfg.MinIO.Bucket)
	}
	if cfg.HTTP.Port != 9090 {
		t.Errorf("HTTP.Port = %d, want 9090 to be preserved", cfg.HTTP.Port)
	}
}

func TestConfig_UnmarshalJSON_PartialBlobThenDefaults(t *testing.T) {
	// A Consul KV blob only needs to specify overrides; everything else
	// should come from applyDefaults after unmarshaling.
	blob := []byte(`{"minio": {"bucket": "prod-photos"}, "postgres": {"max_conns": 25}}`)

	var cfg Config
	if err := json.Unmarshal(blob, &cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg.applyDefaults()

	if cfg.MinIO.Bucket != "prod-photos" {
		t.Errorf("MinIO.Bucket = %q, want prod-photos", cfg.MinIO.Bucket)
	}
	if cfg.Postgres.MaxConns != 25 {
		t.Errorf("Postgres.MaxConns = %d, want 25", cfg.Postgres.MaxConns)
	}
	// untouched fields still get defaults
	if cfg.HTTP.Port != 8080 {
		t.Errorf("HTTP.Port = %d, want default 8080", cfg.HTTP.Port)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("Redis.Addr = %q, want default localhost:6379", cfg.Redis.Addr)
	}
}

func TestPostgresConfig_DSN(t *testing.T) {
	cfg := PostgresConfig{Host: "postgresdb", Port: 80, User: "svc_user", Password: "p@ss/w?rd", DBName: "photo_service", SSLMode: false}

	got := cfg.DSN()
	want := "postgres://svc_user:p%40ss%2Fw%3Frd@postgresdb:80/photo_service?sslmode=disable"
	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestPostgresConfig_DSN_SSLModeRequire(t *testing.T) {
	cfg := PostgresConfig{Host: "postgresdb", Port: 80, User: "u", Password: "p", DBName: "db", SSLMode: true}

	if got := cfg.DSN(); got != "postgres://u:p@postgresdb:80/db?sslmode=require" {
		t.Errorf("DSN() = %q, want sslmode=require", got)
	}
}

func TestReadBootstrap_MissingVarsReturnsError(t *testing.T) {
	t.Setenv("HostConsul", "")
	t.Setenv("KeyConsul", "")
	t.Setenv("ServiceConsul", "")

	dir := t.TempDir()
	missingEnvPath := filepath.Join(dir, ".env")

	if _, err := readBootstrap(missingEnvPath); err == nil {
		t.Fatal("expected an error when HostConsul/KeyConsul/ServiceConsul are unset and no .env file exists")
	}
}

func TestReadBootstrap_ReadsFromEnvFile(t *testing.T) {
	// godotenv.Load never overrides a variable that is already present in
	// the environment, even if empty — so these must be fully unset (not
	// t.Setenv'd to "") for the .env file's values to actually apply.
	for _, name := range []string{"HostConsul", "KeyConsul", "ServiceConsul"} {
		if old, ok := os.LookupEnv(name); ok {
			t.Cleanup(func() { os.Setenv(name, old) })
			os.Unsetenv(name)
		}
	}

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "HostConsul=\"http://localhost:8500\"\nKeyConsul=\"photo_service\"\nServiceConsul=\"consul\"\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test .env: %v", err)
	}

	b, err := readBootstrap(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.HostConsul != "http://localhost:8500" || b.KeyConsul != "photo_service" || b.ServiceConsul != "consul" {
		t.Fatalf("unexpected bootstrap values: %+v", b)
	}
}
