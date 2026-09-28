package application

import (
	"context"
	"sync"
	"time"

	"github.com/JIeeiroSst/manage-service/internal/domain/port"
	"github.com/Nerzal/gocloak/v13"
)

const tokenRefreshMargin = 10 * time.Second

type AdminCredentials struct {
	User     string
	Password string
	Realm    string
}

type KeycloakService struct {
	port.IdentityProvider
	creds AdminCredentials

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func NewKeycloakService(idp port.IdentityProvider, creds AdminCredentials) port.KeycloakUsecase {
	return &KeycloakService{
		IdentityProvider: idp,
		creds:            creds,
	}
}

func (s *KeycloakService) AdminToken(ctx context.Context) (string, error) {
	if s.creds.User == "" || s.creds.Password == "" {
		return "", port.ErrAdminNotConfigured
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && time.Now().Before(s.expiresAt) {
		return s.token, nil
	}

	realm := s.creds.Realm
	if realm == "" {
		realm = "master"
	}
	jwt, err := s.GetToken(ctx, realm, gocloak.TokenOptions{
		ClientID:  gocloak.StringP("admin-cli"),
		GrantType: gocloak.StringP("password"),
		Username:  gocloak.StringP(s.creds.User),
		Password:  gocloak.StringP(s.creds.Password),
	})
	if err != nil {
		return "", err
	}

	s.token = jwt.AccessToken
	s.expiresAt = time.Now().Add(time.Duration(jwt.ExpiresIn)*time.Second - tokenRefreshMargin)
	return s.token, nil
}
