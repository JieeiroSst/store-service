package mfa

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
)

type SessionStarter interface {
	Start(ctx context.Context, userID int, username, provider string) (dto.LoginResponse, error)
}

type Config struct {
	ChallengeTTL time.Duration
}

type Service struct {
	provider   output.MFAProvider
	challenges output.MFAChallengeStore
	sessions   SessionStarter
	tokenGen   output.TokenGenerator
	tokenStore output.TokenStore
	cfg        Config
}

func New(provider output.MFAProvider, challenges output.MFAChallengeStore, sessions SessionStarter, tokenGen output.TokenGenerator, tokenStore output.TokenStore, cfg Config) *Service {
	return &Service{
		provider:   provider,
		challenges: challenges,
		sessions:   sessions,
		tokenGen:   tokenGen,
		tokenStore: tokenStore,
		cfg:        cfg,
	}
}

func (s *Service) Start(ctx context.Context, userID int, username, provider string) (dto.LoginResponse, error) {
	st, err := s.provider.Status(ctx, userID)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if !st.Enabled {
		return s.sessions.Start(ctx, userID, username, provider)
	}
	token, err := newToken()
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if err := s.challenges.Save(ctx, token, domain.MFAChallenge{UserID: userID, Username: username, Provider: provider}, s.cfg.ChallengeTTL); err != nil {
		return dto.LoginResponse{}, err
	}
	return dto.LoginResponse{}, domain.NewMFARequired(token, int64(s.cfg.ChallengeTTL.Seconds()))
}

func (s *Service) VerifyLogin(ctx context.Context, mfaToken, code string) (dto.LoginResponse, error) {
	if mfaToken == "" {
		return dto.LoginResponse{}, domain.ErrMFAChallenge
	}
	ch, err := s.challenges.Get(ctx, mfaToken)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if err := s.provider.VerifyTOTP(ctx, ch.UserID, strings.TrimSpace(code)); err != nil {
		if errors.Is(err, domain.ErrMFANotEnabled) {
			_, _ = s.challenges.Delete(ctx, mfaToken)
			return dto.LoginResponse{}, domain.ErrMFAChallenge
		}
		return dto.LoginResponse{}, err
	}
	deleted, err := s.challenges.Delete(ctx, mfaToken)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if !deleted {
		return dto.LoginResponse{}, domain.ErrMFAChallenge
	}
	return s.sessions.Start(ctx, ch.UserID, ch.Username, ch.Provider)
}

func (s *Service) Status(ctx context.Context, accessToken string) (dto.MFAStatusResponse, error) {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return dto.MFAStatusResponse{}, err
	}
	st, err := s.provider.Status(ctx, cur.UserID)
	if err != nil {
		return dto.MFAStatusResponse{}, err
	}
	if !st.Enabled {
		return dto.MFAStatusResponse{}, nil
	}
	return dto.MFAStatusResponse{
		Enabled:                true,
		Method:                 domain.MFAMethodTOTP,
		EnabledAt:              st.EnabledAt,
		RecoveryCodesRemaining: st.RecoveryCodesRemaining,
	}, nil
}

func (s *Service) SetupTOTP(ctx context.Context, accessToken string) (dto.TOTPSetupResponse, error) {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return dto.TOTPSetupResponse{}, err
	}
	account := cur.Username
	if account == "" {
		account = strconv.Itoa(cur.UserID)
	}
	setup, err := s.provider.SetupTOTP(ctx, cur.UserID, account)
	if err != nil {
		return dto.TOTPSetupResponse{}, err
	}
	return dto.TOTPSetupResponse{
		Secret:     setup.Secret,
		OtpauthURI: setup.OtpauthURI,
		QRCode:     "data:image/png;base64," + base64.StdEncoding.EncodeToString(setup.QRCodePNG),
		Issuer:     setup.Issuer,
		Account:    setup.Account,
		Digits:     setup.Digits,
		Period:     setup.Period,
		Algorithm:  setup.Algorithm,
	}, nil
}

func (s *Service) EnableTOTP(ctx context.Context, accessToken, code string) (dto.RecoveryCodesResponse, error) {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return dto.RecoveryCodesResponse{}, err
	}
	codes, err := s.provider.EnableTOTP(ctx, cur.UserID, strings.TrimSpace(code))
	if err != nil {
		return dto.RecoveryCodesResponse{}, err
	}
	return dto.RecoveryCodesResponse{RecoveryCodes: codes}, nil
}

func (s *Service) DisableTOTP(ctx context.Context, accessToken, code string) error {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return err
	}
	return s.provider.DisableTOTP(ctx, cur.UserID, strings.TrimSpace(code))
}

func (s *Service) RegenerateRecoveryCodes(ctx context.Context, accessToken, code string) (dto.RecoveryCodesResponse, error) {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return dto.RecoveryCodesResponse{}, err
	}
	codes, err := s.provider.RegenerateRecoveryCodes(ctx, cur.UserID, strings.TrimSpace(code))
	if err != nil {
		return dto.RecoveryCodesResponse{}, err
	}
	return dto.RecoveryCodesResponse{RecoveryCodes: codes}, nil
}

func (s *Service) current(ctx context.Context, accessToken string) (domain.Session, error) {
	if _, err := s.tokenGen.ParseAccessToken(ctx, accessToken); err != nil {
		return domain.Session{}, domain.ErrFailedToken
	}
	cur, err := s.tokenStore.GetSessionByAccessToken(ctx, accessToken)
	if err != nil {
		return domain.Session{}, domain.ErrFailedToken
	}
	return cur, nil
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
