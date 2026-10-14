package sso

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/geared_id"
)

const maxProfileBytes = 64 << 10

type SessionStarter interface {
	Start(ctx context.Context, userID int, username, provider string) (dto.LoginResponse, error)
}

type Service struct {
	provider     string
	idp          output.IdentityProvider
	userRepo     output.UserRepository
	socialLogins output.SocialLoginRepository
	hasher       output.Hasher
	authorizer   output.Authorizer
	sessions     SessionStarter
}

func New(provider string, idp output.IdentityProvider, userRepo output.UserRepository, socialLogins output.SocialLoginRepository, hasher output.Hasher, authorizer output.Authorizer, sessions SessionStarter) *Service {
	return &Service{provider: provider, idp: idp, userRepo: userRepo, socialLogins: socialLogins, hasher: hasher, authorizer: authorizer, sessions: sessions}
}

func (s *Service) LoginURL(ctx context.Context, state, verifier string) (string, error) {
	url, err := s.idp.AuthCodeURL(ctx, state, verifier)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrServiceUnavailable, err)
	}
	return url, nil
}

func (s *Service) Callback(ctx context.Context, code, verifier string) (dto.LoginResponse, error) {
	if code == "" {
		return dto.LoginResponse{}, domain.ErrInvalidRequest
	}
	identity, err := s.idp.Exchange(ctx, code, verifier)
	if err != nil {
		log.Printf("sso %s: exchange code: %v", s.provider, err)
		return dto.LoginResponse{}, domain.ErrSSOExchangeFailed
	}
	return s.login(ctx, identity)
}

func (s *Service) LoginWithToken(ctx context.Context, token string) (dto.LoginResponse, error) {
	if token == "" {
		return dto.LoginResponse{}, domain.ErrInvalidRequest
	}
	identity, err := s.idp.VerifyToken(ctx, token)
	if errors.Is(err, domain.ErrSSOTokenUnsupported) {
		return dto.LoginResponse{}, err
	}
	if err != nil {
		log.Printf("sso %s: verify token: %v", s.provider, err)
		return dto.LoginResponse{}, domain.ErrSSOIDTokenInvalid
	}
	return s.login(ctx, identity)
}

func (s *Service) login(ctx context.Context, identity domain.ExternalIdentity) (dto.LoginResponse, error) {
	if identity.Subject == "" {
		return dto.LoginResponse{}, domain.ErrSSONoIdentity
	}
	user, err := s.resolve(ctx, identity)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if !user.Checked {
		return dto.LoginResponse{}, domain.ErrAccountLocked
	}
	return s.sessions.Start(ctx, user.Id, user.Username, s.provider)
}

func (s *Service) resolve(ctx context.Context, identity domain.ExternalIdentity) (domain.User, error) {
	linked, err := s.socialLogins.Find(ctx, s.provider, identity.Subject)
	if err == nil {
		return s.userRepo.FindUser(ctx, linked.UserID)
	}
	if !errors.Is(err, domain.ErrUserNotExist) {
		return domain.User{}, err
	}

	email := ""
	if identity.EmailVerified {
		email = strings.TrimSpace(identity.Email)
	}
	if email != "" {
		users, _, err := s.userRepo.SearchUsers(ctx, domain.UserFilter{Email: email, Limit: 1})
		if err != nil {
			return domain.User{}, err
		}
		if len(users) > 0 {
			if err := s.link(ctx, users[0].Id, identity); err != nil {
				return domain.User{}, err
			}
			return users[0], nil
		}
	}

	created, err := s.create(ctx, identity, email)
	if err != nil {
		return domain.User{}, err
	}
	if err := s.link(ctx, created.Id, identity); err != nil {
		s.rollback(ctx, created.Id)
		return domain.User{}, err
	}
	return created, nil
}

func (s *Service) link(ctx context.Context, userID int, identity domain.ExternalIdentity) error {
	profile := "{}"
	if json.Valid(identity.Profile) && len(identity.Profile) <= maxProfileBytes {
		profile = string(identity.Profile)
	}
	return s.socialLogins.Create(ctx, domain.SocialLogin{
		UserID:        userID,
		Provider:      s.provider,
		Subject:       identity.Subject,
		Email:         identity.Email,
		EmailVerified: identity.EmailVerified,
		Name:          identity.Name,
		Username:      identity.Username,
		AvatarURL:     identity.AvatarURL,
		Scopes:        strings.Join(identity.Scopes, " "),
		Profile:       profile,
		ConsentIP:     domain.ClientInfoFrom(ctx).IP,
		ConsentedAt:   time.Now(),
	})
}

func (s *Service) create(ctx context.Context, identity domain.ExternalIdentity, email string) (domain.User, error) {
	username, err := s.freeUsername(ctx, identity.Username, email, s.provider+"_"+identity.Subject)
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
		Email:      email,
		Name:       identity.Name,
		Checked:    true,
		CreateTime: time.Now(),
	})
	if err != nil {
		return domain.User{}, err
	}

	if _, err := s.authorizer.AssignRoles(ctx, created.Id, domain.DefaultRole); err != nil {
		s.rollback(ctx, created.Id)
		return domain.User{}, fmt.Errorf("%w: %v", domain.ErrAssignRoleFailed, err)
	}
	return created, nil
}

func (s *Service) rollback(ctx context.Context, userID int) {
	if err := s.userRepo.DeleteAccount(ctx, userID); err != nil {
		log.Printf("sso %s: rollback account %d failed: %v", s.provider, userID, err)
	}
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
