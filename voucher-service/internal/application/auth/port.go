package auth

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/user"
)

type RegisterInput struct {
	Email       string
	Password    string
	Role        user.Role
	CorporateID *shared.CorporateID
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (*user.User, error)
	Login(ctx context.Context, in LoginInput) (*LoginOutput, error)
	VerifyToken(ctx context.Context, token string) (*Claims, error)
}

type Claims struct {
	UserID      string
	Role        string
	CorporateID string
}

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id shared.UserID) (*user.User, error)
}

type TokenIssuer interface {
	Issue(claims Claims) (token string, expiresIn int64, err error)
	Verify(token string) (*Claims, error)
}
