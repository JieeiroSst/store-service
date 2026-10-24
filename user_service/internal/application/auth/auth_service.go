package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/copy"
)

type Sessions interface {
	Start(ctx context.Context, userID int, username, provider string) (dto.LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (dto.RefreshResponse, error)
	End(ctx context.Context, accessToken string) error
}

type Service struct {
	userRepo   output.UserRepository
	hasher     output.Hasher
	tokenGen   output.TokenGenerator
	tokenStore output.TokenStore
	sessions   Sessions
	throttle   output.LoginThrottle

	dummyOnce sync.Once
	dummyHash string
}

func New(userRepo output.UserRepository, hasher output.Hasher, tokenGen output.TokenGenerator, tokenStore output.TokenStore, sessions Sessions, throttle output.LoginThrottle) *Service {
	return &Service{
		userRepo:   userRepo,
		hasher:     hasher,
		tokenGen:   tokenGen,
		tokenStore: tokenStore,
		sessions:   sessions,
		throttle:   throttle,
	}
}

func (s *Service) burnPasswordCheck(password string) {
	s.dummyOnce.Do(func() {
		s.dummyHash, _ = s.hasher.HashPassword("user-service-timing-equalizer")
	})
	if s.dummyHash != "" {
		_ = s.hasher.CheckPassword(password, s.dummyHash)
	}
}

func (s *Service) rejectCredentials(ctx context.Context, username string) error {
	if err := s.throttle.RecordFailure(ctx, username); err != nil {
		return err
	}
	return domain.ErrInvalidCredentials
}

func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	var user domain.User
	if err := copy.CopyObject(&req, &user); err != nil {
		return dto.LoginResponse{}, err
	}

	blocked, err := s.throttle.Blocked(ctx, user.Username)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if blocked {
		return dto.LoginResponse{}, domain.ErrTooManyAttempts
	}

	userID, hashedPassword, _, err := s.userRepo.CheckAccount(ctx, user)
	if errors.Is(err, domain.ErrUserNotExist) {
		s.burnPasswordCheck(user.Password)
		return dto.LoginResponse{}, s.rejectCredentials(ctx, user.Username)
	}
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if err := s.hasher.CheckPassword(user.Password, hashedPassword); err != nil {
		return dto.LoginResponse{}, s.rejectCredentials(ctx, user.Username)
	}
	_ = s.throttle.Reset(ctx, user.Username)
	account, err := s.userRepo.FindUser(ctx, userID)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if !account.Checked {
		return dto.LoginResponse{}, domain.ErrAccountLocked
	}

	return s.sessions.Start(ctx, userID, user.Username, domain.ProviderPassword)
}

func (s *Service) Logout(ctx context.Context, req dto.LogoutRequest) (dto.LogoutResponse, error) {
	if err := s.sessions.End(ctx, req.SessionToken); err != nil {
		return dto.LogoutResponse{Message: "false"}, nil
	}
	return dto.LogoutResponse{Message: "true"}, nil
}

func (s *Service) ValidateSession(ctx context.Context, req dto.ValidateRequest) (dto.ValidateResponse, error) {
	claims, err := s.tokenGen.ParseAccessToken(ctx, req.SessionToken)
	if err != nil {
		return dto.ValidateResponse{Valid: false}, nil
	}
	if _, err := s.tokenStore.GetSessionByAccessToken(ctx, req.SessionToken); err != nil {
		return dto.ValidateResponse{Valid: false}, nil
	}

	return dto.ValidateResponse{
		Valid:  true,
		UserId: strconv.Itoa(claims.UserID),
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req dto.RefreshRequest) (dto.RefreshResponse, error) {
	return s.sessions.Refresh(ctx, req.RefreshToken)
}

func (s *Service) Authentication(ctx context.Context, req dto.AuthenticationRequest) (dto.AuthenticationResponse, error) {
	parts := strings.Split(req.Token, " ")
	if len(parts) != 2 {
		return dto.AuthenticationResponse{}, domain.ErrFailedToken
	}

	claims, err := s.tokenGen.ParseAccessToken(ctx, parts[1])
	if err != nil {
		return dto.AuthenticationResponse{}, err
	}
	if claims.Username != req.Username {
		return dto.AuthenticationResponse{}, domain.ErrFailedTokenUsername
	}
	if _, err := s.tokenStore.GetSessionByAccessToken(ctx, parts[1]); err != nil {
		return dto.AuthenticationResponse{}, domain.ErrFailedToken
	}

	return dto.AuthenticationResponse{Valid: true, Message: "success"}, nil
}
