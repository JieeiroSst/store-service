package auth

import (
	"context"
	"errors"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type AuthService interface {
	VerifyToken(ctx context.Context, token string) (*Claims, error)
}

type Claims struct {
	UserID string
	Role   string
}

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*Claims, error)
}
