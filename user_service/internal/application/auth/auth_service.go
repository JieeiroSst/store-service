package auth

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/copy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	userRepo   output.UserRepository
	authorizer output.Authorizer
	hasher     output.Hasher
	tokenGen   output.TokenGenerator
	tokenStore output.TokenStore
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(userRepo output.UserRepository, authorizer output.Authorizer, hasher output.Hasher, tokenGen output.TokenGenerator, tokenStore output.TokenStore, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		userRepo:   userRepo,
		authorizer: authorizer,
		hasher:     hasher,
		tokenGen:   tokenGen,
		tokenStore: tokenStore,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	var user domain.User
	if err := copy.CopyObject(&req, &user); err != nil {
		return dto.LoginResponse{}, err
	}

	userID, hashedPassword, _, err := s.userRepo.CheckAccount(ctx, user)
	if err != nil {
		return dto.LoginResponse{}, errors.New("user does not exist")
	}
	if err := s.hasher.CheckPassword(user.Password, hashedPassword); err != nil {
		return dto.LoginResponse{}, errors.New("password entered incorrectly")
	}

	pair, err := s.issueTokenPair(ctx, userID, user.Username, s.userRoles(ctx, userID))
	if err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.LoginResponse{
		SessionToken: pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiryTime:   pair.ExpiresIn,
	}, nil
}

func (s *Service) Logout(ctx context.Context, req dto.LogoutRequest) (dto.LogoutResponse, error) {
	session, err := s.tokenStore.GetSessionByAccessToken(ctx, req.SessionToken)
	if err != nil {
		return dto.LogoutResponse{Message: "false"}, nil
	}
	if err := s.tokenStore.DeleteSession(ctx, session.AccessToken, session.RefreshToken); err != nil {
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
	session, err := s.tokenStore.GetSessionByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return dto.RefreshResponse{}, status.Errorf(codes.Unauthenticated, "refresh token is invalid or expired, please login again")
	}

	// Re-read the role so grants/revocations take effect on refresh.
	pair, err := s.issueTokenPair(ctx, session.UserID, session.Username, s.userRoles(ctx, session.UserID))
	if err != nil {
		return dto.RefreshResponse{}, err
	}

	_ = s.tokenStore.DeleteSession(ctx, session.AccessToken, session.RefreshToken)

	return dto.RefreshResponse{
		NewSessionToken: pair.AccessToken,
		NewRefreshToken: pair.RefreshToken,
		ExpiryTime:      pair.ExpiresIn,
	}, nil
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

// userRoles reads the user's roles from authorize-service for the token's
// informational "role" (primary) and "roles" (every effective role) claims.
// It fails open to no roles rather than blocking login: real access
// decisions are made by authorize-service.
func (s *Service) userRoles(ctx context.Context, userID int) domain.UserRoles {
	roles, err := s.authorizer.GetUserRoles(ctx, userID)
	if err != nil {
		log.Printf("auth: get roles for user %d from authorize-service: %v", userID, err)
		return domain.UserRoles{}
	}
	return roles
}

func (s *Service) issueTokenPair(ctx context.Context, userID int, username string, roles domain.UserRoles) (domain.TokenPair, error) {
	role := roles.PrimaryRole
	accessToken, err := s.tokenGen.GenerateAccessToken(ctx, userID, username, role, roles.EffectiveRoles)
	if err != nil {
		return domain.TokenPair{}, err
	}
	refreshToken, err := s.tokenGen.GenerateRefreshToken(ctx)
	if err != nil {
		return domain.TokenPair{}, err
	}

	session := domain.Session{
		UserID:       userID,
		Username:     username,
		Role:         role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if err := s.tokenStore.SaveSession(ctx, session, s.accessTTL, s.refreshTTL); err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}
