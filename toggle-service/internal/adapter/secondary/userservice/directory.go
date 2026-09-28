package userservice

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

// directory adapts Client to port.UserDirectory.
type directory struct{ c *Client }

func NewDirectory(cfg *config.Config) port.UserDirectory {
	return directory{c: New(cfg.UserService.BaseURL, cfg.UserService.Timeout)}
}

func (d directory) Authenticate(ctx context.Context, token string) (*model.User, error) {
	id, err := d.c.Authenticate(ctx, token)
	if errors.Is(err, ErrUnauthenticated) {
		return nil, apperr.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	return &model.User{ID: id.UserID, Username: id.Username, IsAdmin: id.IsAdmin()}, nil
}
