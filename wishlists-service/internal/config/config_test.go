package config

import (
	"net/url"
	"testing"
)

func TestDSNEscapesCredentials(t *testing.T) {
	p := PostgresConfig{Host: "db", Port: 5432, User: "svc", Password: "p@ss word/='x", Database: "wishlists_service", SSLMode: "disable"}
	u, err := url.Parse(p.DSN())
	if err != nil {
		t.Fatal(err)
	}
	pw, _ := u.User.Password()
	if pw != p.Password || u.User.Username() != "svc" || u.Host != "db:5432" || u.Path != "/wishlists_service" || u.Query().Get("sslmode") != "disable" {
		t.Fatalf("unexpected dsn %q", p.DSN())
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	t.Setenv("CACHE_USER_TTL", "-1s")
	if _, err := Load(); err == nil {
		t.Fatal("negative ttl must be rejected")
	}
	t.Setenv("CACHE_USER_TTL", "")
	t.Setenv("HTTP_PORT", "abc")
	if _, err := Load(); err == nil {
		t.Fatal("bad port must be rejected")
	}
}
