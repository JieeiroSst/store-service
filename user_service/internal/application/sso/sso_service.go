package sso

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/geared_id"
)

type SessionIssuer interface {
	IssueSession(ctx context.Context, userID int, username string) (dto.LoginResponse, error)
}

type Service struct {
	idp        output.IdentityProvider
	userRepo   output.UserRepository
	hasher     output.Hasher
	authorizer output.Authorizer
	sessions   SessionIssuer
}

func New(idp output.IdentityProvider, userRepo output.UserRepository, hasher output.Hasher, authorizer output.Authorizer, sessions SessionIssuer) *Service {
	return &Service{idp: idp, userRepo: userRepo, hasher: hasher, authorizer: authorizer, sessions: sessions}
}

func (s *Service) LoginURL(ctx context.Context, state string) (string, error) {
	url, err := s.idp.AuthCodeURL(ctx, state)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrServiceUnavailable, err)
	}
	return url, nil
}

func (s *Service) Callback(ctx context.Context, code string) (dto.LoginResponse, error) {
	if code == "" {
		return dto.LoginResponse{}, domain.ErrInvalidRequest
	}
	identity, err := s.idp.Exchange(ctx, code)
	if err != nil {
		log.Printf("sso: exchange code: %v", err)
		return dto.LoginResponse{}, domain.ErrSSOExchangeFailed
	}
	if identity.Email == "" || !identity.EmailVerified {
		return dto.LoginResponse{}, domain.ErrSSOEmailUnverified
	}

	user, err := s.findOrCreate(ctx, identity)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if !user.Checked {
		return dto.LoginResponse{}, domain.ErrAccountLocked
	}
	return s.sessions.IssueSession(ctx, user.Id, user.Username)
}

func (s *Service) findOrCreate(ctx context.Context, identity domain.ExternalIdentity) (domain.User, error) {
	users, _, err := s.userRepo.SearchUsers(ctx, domain.UserFilter{Email: identity.Email, Limit: 1})
	if err != nil {
		return domain.User{}, err
	}
	if len(users) > 0 {
		return users[0], nil
	}

	username, err := s.freeUsername(ctx, identity.Username, identity.Email)
	if err != nil {
		return domain.User{}, err
	}
	password, err := randomPassword()
	if err != nil {
		return domain.User{}, err
	}
	hashedPassword, err := s.hasher.HashPassword(password)
	if err != nil {
		return domain.User{}, domain.ErrHashPasswordFailed
	}

	created, err := s.userRepo.CreateAccount(ctx, domain.User{
		Id:         geared_id.GearedIntID(),
		Username:   username,
		Password:   hashedPassword,
		Email:      identity.Email,
		Name:       identity.Name,
		Checked:    true,
		CreateTime: time.Now(),
	})
	if err != nil {
		return domain.User{}, err
	}

	if _, err := s.authorizer.AssignRoles(ctx, created.Id, domain.DefaultRole); err != nil {
		if delErr := s.userRepo.DeleteAccount(ctx, created.Id); delErr != nil {
			log.Printf("sso: rollback account %d failed: %v", created.Id, delErr)
		}
		return domain.User{}, fmt.Errorf("%w: %v", domain.ErrAssignRoleFailed, err)
	}
	return created, nil
}

func (s *Service) freeUsername(ctx context.Context, candidates ...string) (string, error) {
	for _, name := range candidates {
		if name == "" {
			continue
		}
		err := s.userRepo.CheckAccountExists(ctx, domain.User{Username: name})
		if err == nil {
			return name, nil
		}
		if !errors.Is(err, domain.ErrUserExist) {
			return "", err
		}
	}
	return "", domain.ErrUserExist
}

func randomPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
