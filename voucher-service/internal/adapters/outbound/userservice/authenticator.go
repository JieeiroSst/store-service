package userservice

import (
	"context"
	"errors"

	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"go.uber.org/fx"
)

type authenticator struct{ c *Client }

func NewAuthenticator(cfg *config.Config) authapp.Authenticator {
	return authenticator{c: New(cfg.UserServiceURL, cfg.UserServiceTimeout)}
}

func (a authenticator) Authenticate(ctx context.Context, token string) (*authapp.Claims, error) {
	id, err := a.c.Authenticate(ctx, token)
	if errors.Is(err, ErrUnauthenticated) {
		return nil, authapp.ErrUnauthenticated
	}
	if err != nil {
		return nil, err
	}
	return &authapp.Claims{UserID: id.UserID, Role: id.Role}, nil
}

var Module = fx.Module("userservice",
	fx.Provide(NewAuthenticator),
)
