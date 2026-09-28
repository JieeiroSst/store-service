package auth

import (
	"context"

	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

// service authenticates admin-API callers with user-service tokens. Sign-up
// and login happen in user-service only; this service issues no tokens.
type service struct {
	users port.UserDirectory
}

func NewService(users port.UserDirectory) port.AuthService {
	return &service{users: users}
}

func (s *service) VerifyToken(ctx context.Context, tokenString string) (string, bool, error) {
	u, err := s.users.Authenticate(ctx, tokenString)
	if err != nil {
		return "", false, err
	}
	return u.ID, u.IsAdmin, nil
}
