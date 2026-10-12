package userservice

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary/userservice/userservicetest"
	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

func newClient(t *testing.T, jidField string) (*Client, *userservicetest.Fake) {
	t.Helper()
	fake := userservicetest.NewFake("internal-secret",
		userservicetest.FakeUser{ID: 7, Username: "alice", Password: "alicepw"},
		userservicetest.FakeUser{ID: 8, Username: "bob", Password: "bobpw"},
	)
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	c := New(config.UserServiceConfig{
		URL: srv.URL, Timeout: 2 * time.Second, InternalToken: "internal-secret",
		JIDField: jidField, AllowPasswordLogin: true, CacheTTL: time.Minute,
	})
	return c, fake
}

func TestCheckPasswordWithToken(t *testing.T) {
	ctx := context.Background()
	c, fake := newClient(t, "username")
	tok, _ := fake.Login("alice", "alicepw")
	if ok, err := c.CheckPassword(ctx, "alice", "alice", "localhost", tok); err != nil || !ok {
		t.Fatalf("valid token rejected: %v %v", ok, err)
	}
	if ok, _ := c.CheckPassword(ctx, "bob", "bob", "localhost", tok); ok {
		t.Fatal("alice's token must not log in as bob")
	}
	fake.Revoke(tok)
	if ok, _ := c.CheckPassword(ctx, "alice", "alice", "localhost", tok); ok {
		t.Fatal("revoked token accepted")
	}
}

func TestCheckPasswordWithPassword(t *testing.T) {
	ctx := context.Background()
	c, fake := newClient(t, "username")
	if ok, err := c.CheckPassword(ctx, "alice", "alice", "localhost", "alicepw"); err != nil || !ok {
		t.Fatalf("password login rejected: %v %v", ok, err)
	}
	if fake.Logouts.Load() != 1 {
		t.Fatalf("password login must close the user_service session, logouts=%d", fake.Logouts.Load())
	}
	if ok, _ := c.CheckPassword(ctx, "alice", "alice", "localhost", "wrong"); ok {
		t.Fatal("wrong password accepted")
	}
	c.allowPassword = false
	if ok, _ := c.CheckPassword(ctx, "alice", "alice", "localhost", "alicepw"); ok {
		t.Fatal("password login must be refused when allow_password_login is false")
	}
}

func TestUserIDMode(t *testing.T) {
	ctx := context.Background()
	c, fake := newClient(t, "user_id")
	tok, _ := fake.Login("bob", "bobpw")
	if ok, _ := c.CheckPassword(ctx, "8", "8", "localhost", tok); !ok {
		t.Fatal("token must log in as JID 8@host")
	}
	if ok, _ := c.CheckPassword(ctx, "bob", "bob", "localhost", tok); ok {
		t.Fatal("user_id mode must not accept the username as JID")
	}
	if ok, _ := c.CheckPassword(ctx, "8", "8", "localhost", "bobpw"); ok {
		t.Fatal("user_id mode has no password login")
	}
	if ok, err := c.UserExists(ctx, "8", "localhost"); err != nil || !ok {
		t.Fatalf("user 8 exists: %v %v", ok, err)
	}
	if ok, _ := c.UserExists(ctx, "99", "localhost"); ok {
		t.Fatal("user 99 must not exist")
	}
}

func TestUserExistsIsCached(t *testing.T) {
	ctx := context.Background()
	c, fake := newClient(t, "username")
	for i := 0; i < 5; i++ {
		if ok, err := c.UserExists(ctx, "alice", "localhost"); err != nil || !ok {
			t.Fatalf("alice exists: %v %v", ok, err)
		}
		if ok, _ := c.UserExists(ctx, "nobody", "localhost"); ok {
			t.Fatal("nobody must not exist")
		}
	}
	if n := fake.Lookups.Load(); n != 2 {
		t.Fatalf("lookups = %d, want 2 (cached)", n)
	}
}

func TestUpstreamDown(t *testing.T) {
	c := New(config.UserServiceConfig{URL: "http://127.0.0.1:1", Timeout: 200 * time.Millisecond, JIDField: "username", AllowPasswordLogin: true})
	if ok, err := c.CheckPassword(context.Background(), "alice", "alice", "localhost", "pw"); ok || err == nil {
		t.Fatalf("expected upstream error, got %v %v", ok, err)
	}
	if _, err := c.UserExists(context.Background(), "alice", "localhost"); err == nil {
		t.Fatal("expected upstream error")
	}
}

func TestEnsureUser(t *testing.T) {
	ctx := context.Background()
	c, fake := newClient(t, "username")
	if ok, _ := c.UserExists(ctx, "admin", "localhost"); ok {
		t.Fatal("admin must not exist yet")
	}
	created, err := c.EnsureUser(ctx, domain.Account{Username: "admin", Password: "Admin123", Email: "ejabberd.admin@store.com"})
	if err != nil || !created {
		t.Fatalf("create admin: %v %v", created, err)
	}
	if ok, _ := c.UserExists(ctx, "admin", "localhost"); !ok {
		t.Fatal("negative cache must be dropped after sign-up")
	}
	if ok, _ := c.CheckPassword(ctx, "admin", "admin", "localhost", "Admin123"); !ok {
		t.Fatal("created admin cannot log in")
	}
	created, err = c.EnsureUser(ctx, domain.Account{Username: "admin", Password: "Admin123", Email: "ejabberd.admin@store.com"})
	if err != nil || created {
		t.Fatalf("existing admin: %v %v", created, err)
	}
	_, err = c.EnsureUser(ctx, domain.Account{Username: "weak", Password: "weak", Email: "ejabberd.admin@store.com"})
	if !errors.Is(err, domain.ErrAccountRejected) {
		t.Fatalf("weak password err = %v", err)
	}
	fake.Down.Store(1)
	_, err = c.EnsureUser(ctx, domain.Account{Username: "x", Password: "Xpassword1", Email: "ejabberd.admin@store.com"})
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("outage err = %v", err)
	}
}
