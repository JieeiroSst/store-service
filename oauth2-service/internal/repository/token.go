package repository

import (
	"context"

	"github.com/JIeeiroSst/oauth2-service/pkg/token"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ITokensStore interface {
	Create(ctx context.Context, info token.TokenInfo) error

	RemoveByCode(ctx context.Context, code string) error

	RemoveByAccess(ctx context.Context, access string) error

	RemoveByRefresh(ctx context.Context, refresh string) error

	GetByCode(ctx context.Context, code string) (token.TokenInfo, error)

	GetByAccess(ctx context.Context, access string) (token.TokenInfo, error)

	GetByRefresh(ctx context.Context, refresh string) (token.TokenInfo, error)
}

type TokenStore struct {
	db     *gorm.DB
	resdis *redis.Client
}

func NewMemoryTokenStore(db *gorm.DB, resdis *redis.Client) (ITokensStore, error) {
	return &TokenStore{db: db, resdis: resdis}, nil
}

func (ts *TokenStore) Create(ctx context.Context, info token.TokenInfo) error {
	return nil
}

func (ts *TokenStore) RemoveByCode(ctx context.Context, code string) error {
	return nil
}

func (ts *TokenStore) RemoveByAccess(ctx context.Context, access string) error {
	return nil
}

func (ts *TokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	return nil
}

func (ts *TokenStore) GetByCode(ctx context.Context, code string) (token.TokenInfo, error) {
	return nil, nil
}

func (ts *TokenStore) GetByAccess(ctx context.Context, access string) (token.TokenInfo, error) {
	return nil, nil
}

func (ts *TokenStore) GetByRefresh(ctx context.Context, refresh string) (token.TokenInfo, error) {
	return nil, nil
}
