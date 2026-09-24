package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
)

type AuthService struct {
	users port.UserDirectory
}

func NewAuthService(users port.UserDirectory) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (model.Session, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return model.Session{}, port.ErrInvalidInput
	}
	return s.users.Login(ctx, username, password)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (model.Session, error) {
	if refreshToken == "" {
		return model.Session{}, port.ErrInvalidInput
	}
	return s.users.Refresh(ctx, refreshToken)
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (model.User, error) {
	if token == "" {
		return model.User{}, port.ErrUnauthenticated
	}
	return s.users.Validate(ctx, token)
}
