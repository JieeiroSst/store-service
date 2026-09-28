package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/manage-service/internal/domain/port"
	"github.com/Nerzal/gocloak/v13"
)

type fakeIdentityProvider struct {
	port.IdentityProvider
	calls     int
	expiresIn int
	lastRealm string
	lastUser  string
}

func (f *fakeIdentityProvider) GetToken(_ context.Context, realm string, options gocloak.TokenOptions) (*gocloak.JWT, error) {
	f.calls++
	f.lastRealm = realm
	f.lastUser = *options.Username
	return &gocloak.JWT{AccessToken: "token", ExpiresIn: f.expiresIn}, nil
}

func TestAdminTokenNotConfigured(t *testing.T) {
	svc := NewKeycloakService(&fakeIdentityProvider{}, AdminCredentials{})
	if _, err := svc.AdminToken(context.Background()); !errors.Is(err, port.ErrAdminNotConfigured) {
		t.Fatalf("want ErrAdminNotConfigured, got %v", err)
	}
}

func TestAdminTokenIsCached(t *testing.T) {
	idp := &fakeIdentityProvider{expiresIn: 300}
	svc := NewKeycloakService(idp, AdminCredentials{User: "admin", Password: "pw"})

	for i := 0; i < 3; i++ {
		token, err := svc.AdminToken(context.Background())
		if err != nil || token != "token" {
			t.Fatalf("unexpected result %q, %v", token, err)
		}
	}
	if idp.calls != 1 {
		t.Fatalf("want 1 token request, got %d", idp.calls)
	}
	if idp.lastRealm != "master" || idp.lastUser != "admin" {
		t.Fatalf("unexpected login realm=%s user=%s", idp.lastRealm, idp.lastUser)
	}
}

func TestAdminTokenRefreshesWhenExpired(t *testing.T) {
	idp := &fakeIdentityProvider{expiresIn: 1}
	svc := NewKeycloakService(idp, AdminCredentials{User: "admin", Password: "pw", Realm: "ops"})

	svc.AdminToken(context.Background())
	svc.AdminToken(context.Background())
	if idp.calls != 2 {
		t.Fatalf("want 2 token requests, got %d", idp.calls)
	}
	if idp.lastRealm != "ops" {
		t.Fatalf("want realm ops, got %s", idp.lastRealm)
	}
}
