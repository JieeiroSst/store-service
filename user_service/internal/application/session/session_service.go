package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
)

type Config struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	MaxPerUser int
}

type Service struct {
	repo       output.SessionRepository
	store      output.TokenStore
	tokenGen   output.TokenGenerator
	authorizer output.Authorizer
	cfg        Config
	now        func() time.Time
}

func New(repo output.SessionRepository, store output.TokenStore, tokenGen output.TokenGenerator, authorizer output.Authorizer, cfg Config) *Service {
	return &Service{repo: repo, store: store, tokenGen: tokenGen, authorizer: authorizer, cfg: cfg, now: time.Now}
}

func (s *Service) Start(ctx context.Context, userID int, username, provider string) (dto.LoginResponse, error) {
	id, err := newID()
	if err != nil {
		return dto.LoginResponse{}, err
	}
	now := s.now()
	client := domain.ClientInfoFrom(ctx)
	if err := s.repo.Create(ctx, domain.LoginSession{
		ID:           id,
		UserID:       userID,
		Provider:     provider,
		IP:           truncate(client.IP, 64),
		UserAgent:    truncate(client.UserAgent, 512),
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(s.cfg.RefreshTTL),
	}); err != nil {
		return dto.LoginResponse{}, err
	}

	pair, err := s.issue(ctx, id, userID, username)
	if err != nil {
		_ = s.repo.Revoke(ctx, []string{id}, domain.RevokeLogout, now)
		return dto.LoginResponse{}, err
	}
	s.enforceLimit(ctx, userID)

	return dto.LoginResponse{
		SessionToken: pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiryTime:   pair.ExpiresIn,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (dto.RefreshResponse, error) {
	old, err := s.store.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return dto.RefreshResponse{}, domain.ErrRefreshTokenInvalid
	}

	id := old.ID
	if id == "" {
		return s.migrateLegacy(ctx, old)
	}
	row, err := s.repo.Get(ctx, id)
	if err != nil && !errors.Is(err, domain.ErrSessionNotFound) {
		return dto.RefreshResponse{}, err
	}
	now := s.now()
	if err != nil || row.RevokedAt != nil || !row.ExpiresAt.After(now) {
		_ = s.store.DeleteSession(ctx, id)
		return dto.RefreshResponse{}, domain.ErrRefreshTokenInvalid
	}

	pair, err := s.issue(ctx, id, old.UserID, old.Username)
	if err != nil {
		return dto.RefreshResponse{}, err
	}
	_ = s.store.DeleteTokens(ctx, old.AccessToken, old.RefreshToken)
	if err := s.repo.Touch(ctx, id, now, now.Add(s.cfg.RefreshTTL)); err != nil {
		log.Printf("session: touch %s: %v", id, err)
	}

	return dto.RefreshResponse{
		NewSessionToken: pair.AccessToken,
		NewRefreshToken: pair.RefreshToken,
		ExpiryTime:      pair.ExpiresIn,
	}, nil
}

func (s *Service) migrateLegacy(ctx context.Context, old domain.Session) (dto.RefreshResponse, error) {
	resp, err := s.Start(ctx, old.UserID, old.Username, domain.ProviderPassword)
	if err != nil {
		return dto.RefreshResponse{}, err
	}
	_ = s.store.DeleteTokens(ctx, old.AccessToken, old.RefreshToken)
	return dto.RefreshResponse{
		NewSessionToken: resp.SessionToken,
		NewRefreshToken: resp.RefreshToken,
		ExpiryTime:      resp.ExpiryTime,
	}, nil
}

func (s *Service) End(ctx context.Context, accessToken string) error {
	cur, err := s.store.GetSessionByAccessToken(ctx, accessToken)
	if err != nil {
		return domain.ErrFailedToken
	}
	_ = s.store.DeleteTokens(ctx, cur.AccessToken, cur.RefreshToken)
	if cur.ID == "" {
		return nil
	}
	return s.revoke(ctx, []string{cur.ID}, domain.RevokeLogout)
}

func (s *Service) List(ctx context.Context, accessToken string) (dto.ListSessionsResponse, error) {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return dto.ListSessionsResponse{}, err
	}
	rows, err := s.repo.ListActive(ctx, cur.UserID, s.now())
	if err != nil {
		return dto.ListSessionsResponse{}, err
	}
	out := dto.ListSessionsResponse{Sessions: make([]dto.SessionInfo, 0, len(rows))}
	for _, r := range rows {
		out.Sessions = append(out.Sessions, dto.SessionInfo{
			ID:           r.ID,
			Provider:     r.Provider,
			IP:           r.IP,
			UserAgent:    r.UserAgent,
			CreatedAt:    r.CreatedAt,
			LastActiveAt: r.LastActiveAt,
			ExpiresAt:    r.ExpiresAt,
			Current:      r.ID == cur.ID,
		})
	}
	return out, nil
}

func (s *Service) Revoke(ctx context.Context, accessToken, sessionID string) error {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return err
	}
	row, err := s.repo.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if row.UserID != cur.UserID || row.RevokedAt != nil {
		return domain.ErrSessionNotFound
	}
	return s.revoke(ctx, []string{sessionID}, domain.RevokeByUser)
}

func (s *Service) RevokeOthers(ctx context.Context, accessToken string, includeCurrent bool) (dto.RevokeSessionsResponse, error) {
	cur, err := s.current(ctx, accessToken)
	if err != nil {
		return dto.RevokeSessionsResponse{}, err
	}
	rows, err := s.repo.ListActive(ctx, cur.UserID, s.now())
	if err != nil {
		return dto.RevokeSessionsResponse{}, err
	}
	var ids []string
	for _, r := range rows {
		if includeCurrent || r.ID != cur.ID {
			ids = append(ids, r.ID)
		}
	}
	if err := s.revoke(ctx, ids, domain.RevokeByUser); err != nil {
		return dto.RevokeSessionsResponse{}, err
	}
	return dto.RevokeSessionsResponse{Revoked: len(ids)}, nil
}

func (s *Service) RevokeAll(ctx context.Context, userID int, reason string) error {
	rows, err := s.repo.ListActive(ctx, userID, s.now())
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	return s.revoke(ctx, ids, reason)
}

func (s *Service) current(ctx context.Context, accessToken string) (domain.Session, error) {
	if _, err := s.tokenGen.ParseAccessToken(ctx, accessToken); err != nil {
		return domain.Session{}, domain.ErrFailedToken
	}
	cur, err := s.store.GetSessionByAccessToken(ctx, accessToken)
	if err != nil || cur.ID == "" {
		return domain.Session{}, domain.ErrFailedToken
	}
	return cur, nil
}

func (s *Service) revoke(ctx context.Context, ids []string, reason string) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.repo.Revoke(ctx, ids, reason, s.now()); err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.store.DeleteSession(ctx, id); err != nil {
			log.Printf("session: delete tokens of %s: %v", id, err)
		}
	}
	return nil
}

func (s *Service) enforceLimit(ctx context.Context, userID int) {
	if s.cfg.MaxPerUser <= 0 {
		return
	}
	rows, err := s.repo.ListActive(ctx, userID, s.now())
	if err != nil || len(rows) <= s.cfg.MaxPerUser {
		return
	}
	var ids []string
	for _, r := range rows[s.cfg.MaxPerUser:] {
		ids = append(ids, r.ID)
	}
	if err := s.revoke(ctx, ids, domain.RevokeEvicted); err != nil {
		log.Printf("session: evict for user %d: %v", userID, err)
	}
}

func (s *Service) issue(ctx context.Context, id string, userID int, username string) (domain.TokenPair, error) {
	roles := s.userRoles(ctx, userID)
	accessToken, err := s.tokenGen.GenerateAccessToken(ctx, userID, username, roles.PrimaryRole, roles.EffectiveRoles)
	if err != nil {
		return domain.TokenPair{}, err
	}
	refreshToken, err := s.tokenGen.GenerateRefreshToken(ctx)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if err := s.store.SaveSession(ctx, domain.Session{
		ID:           id,
		UserID:       userID,
		Username:     username,
		Role:         roles.PrimaryRole,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, s.cfg.AccessTTL, s.cfg.RefreshTTL); err != nil {
		return domain.TokenPair{}, err
	}
	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTTL.Seconds()),
	}, nil
}

func (s *Service) userRoles(ctx context.Context, userID int) domain.UserRoles {
	roles, err := s.authorizer.GetUserRoles(ctx, userID)
	if err != nil {
		log.Printf("session: get roles for user %d from authorize-service: %v", userID, err)
		return domain.UserRoles{}
	}
	return roles
}

func newID() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
