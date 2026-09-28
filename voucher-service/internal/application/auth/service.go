package auth

import "context"

type Service struct {
	users Authenticator
}

func NewService(users Authenticator) AuthService {
	return &Service{users: users}
}

func (s *Service) VerifyToken(ctx context.Context, token string) (*Claims, error) {
	return s.users.Authenticate(ctx, token)
}
