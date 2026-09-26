package service

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
)

type AuthService struct {
	identity  outbound.IdentityProvider
	adminRole string
}

var _ inbound.AuthUseCase = (*AuthService)(nil)

func NewAuthService(identity outbound.IdentityProvider, adminRole string) *AuthService {
	return &AuthService{identity: identity, adminRole: adminRole}
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (inbound.Principal, error) {
	id, err := s.identity.Resolve(ctx, token)
	if err != nil {
		return inbound.Principal{}, err
	}
	p := inbound.Principal{UserID: id.UserID, Email: id.Email}
	for _, r := range id.Roles {
		if strings.EqualFold(r, s.adminRole) {
			p.Admin = true
			break
		}
	}
	return p, nil
}
