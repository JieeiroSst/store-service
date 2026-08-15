package auth

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/user"
)

type Service struct {
	repo   UserRepository
	tokens TokenIssuer
	clock  shared.Clock
}

func NewService(repo UserRepository, tokens TokenIssuer, clock shared.Clock) AuthService {
	return &Service{repo: repo, tokens: tokens, clock: clock}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*user.User, error) {
	u, err := user.NewUser(in.Email, in.Password, in.Role, s.clock.Now())
	if err != nil {
		return nil, err
	}
	u.CorporateID = in.CorporateID
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	u, err := s.repo.FindByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if !u.IsActive() {
		return nil, user.ErrUserInactive
	}
	if err := u.VerifyPassword(in.Password); err != nil {
		return nil, err
	}

	claims := Claims{UserID: u.ID.String(), Role: string(u.Role)}
	if u.CorporateID != nil {
		claims.CorporateID = u.CorporateID.String()
	}
	token, expiresIn, err := s.tokens.Issue(claims)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{Token: token, ExpiresIn: expiresIn, UserID: claims.UserID, Role: claims.Role}, nil
}

func (s *Service) VerifyToken(ctx context.Context, token string) (*Claims, error) {
	return s.tokens.Verify(token)
}
