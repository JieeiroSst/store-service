package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

type service struct {
	users port.UserRepository
	cfg   *config.Config
}

func NewService(users port.UserRepository, cfg *config.Config) port.AuthService {
	return &service{users: users, cfg: cfg}
}

type claims struct {
	UserID  uuid.UUID `json:"userId"`
	IsAdmin bool      `json:"isAdmin"`
	jwt.RegisteredClaims
}

func (s *service) Register(ctx context.Context, email, username, password string) (*model.User, error) {
	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperr.ErrConflict
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{Email: email, Username: username, PasswordHash: string(hash)}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, apperr.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, apperr.ErrInvalidCredentials
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
			Subject:   u.ID.String(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *service) VerifyToken(ctx context.Context, tokenString string) (uuid.UUID, bool, error) {
	var c claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, false, apperr.ErrUnauthorized
	}
	return c.UserID, c.IsAdmin, nil
}
