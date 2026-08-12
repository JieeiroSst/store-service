package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

type service struct {
	users port.UserDirectory
	cfg   *config.Config
}

func NewService(users port.UserDirectory, cfg *config.Config) port.AuthService {
	return &service{users: users, cfg: cfg}
}

type claims struct {
	UserID  string `json:"userId"`
	IsAdmin bool   `json:"isAdmin"`
	jwt.RegisteredClaims
}

func (s *service) Register(ctx context.Context, email, username, password string) (*model.User, error) {
	return s.users.Register(ctx, email, username, password)
}

func (s *service) Login(ctx context.Context, username, password string) (string, *model.User, error) {
	u, err := s.users.Login(ctx, username, password)
	if err != nil {
		return "", nil, err
	}

	token, err := s.issueToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *service) issueToken(u *model.User) (string, error) {
	now := time.Now()
	c := claims{
		UserID:  u.ID,
		IsAdmin: u.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.cfg.JWT.ExpiryMinutes) * time.Minute)),
			Subject:   u.ID,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *service) VerifyToken(ctx context.Context, tokenString string) (string, bool, error) {
	var c claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return "", false, apperr.ErrUnauthorized
	}
	return c.UserID, c.IsAdmin, nil
}
