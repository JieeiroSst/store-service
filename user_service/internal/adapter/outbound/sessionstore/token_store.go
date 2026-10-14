package sessionstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/utils/cache/expire"
)

const (
	accessKeyPrefix  = "user_service:session:access:"
	refreshKeyPrefix = "user_service:session:refresh:"
	idKeyPrefix      = "user_service:session:id:"
)

type TokenStore struct {
	cache expire.CacheHelper
}

func New(cache expire.CacheHelper) *TokenStore {
	return &TokenStore{cache: cache}
}

func (t *TokenStore) SaveSession(ctx context.Context, session domain.Session, accessTTL, refreshTTL time.Duration) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return err
	}
	if err := t.cache.SetInterface(ctx, accessKeyPrefix+session.AccessToken, string(payload), accessTTL); err != nil {
		return err
	}
	if err := t.cache.SetInterface(ctx, refreshKeyPrefix+session.RefreshToken, string(payload), refreshTTL); err != nil {
		return err
	}
	if session.ID != "" {
		if err := t.cache.SetInterface(ctx, idKeyPrefix+session.ID, string(payload), refreshTTL); err != nil {
			return err
		}
	}
	return nil
}

func (t *TokenStore) GetSessionByAccessToken(ctx context.Context, accessToken string) (domain.Session, error) {
	return t.get(ctx, accessKeyPrefix+accessToken)
}

func (t *TokenStore) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (domain.Session, error) {
	return t.get(ctx, refreshKeyPrefix+refreshToken)
}

func (t *TokenStore) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	session, err := t.get(ctx, idKeyPrefix+sessionID)
	if err == nil {
		_ = t.DeleteTokens(ctx, session.AccessToken, session.RefreshToken)
	}
	return t.cache.Removekey(ctx, idKeyPrefix+sessionID)
}

func (t *TokenStore) DeleteTokens(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		_ = t.cache.Removekey(ctx, accessKeyPrefix+accessToken)
	}
	if refreshToken != "" {
		_ = t.cache.Removekey(ctx, refreshKeyPrefix+refreshToken)
	}
	return nil
}

func (t *TokenStore) get(ctx context.Context, key string) (domain.Session, error) {
	cached, err := t.cache.GetInterface(ctx, key)
	if err != nil {
		return domain.Session{}, err
	}
	raw, ok := cached.(string)
	if !ok {
		return domain.Session{}, domain.ErrFailedToken
	}
	var session domain.Session
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return domain.Session{}, err
	}
	return session, nil
}
