package repository

import (
	"context"

	"github.com/JIeeiroSst/oauth2-service/pkg/token"
	"gorm.io/gorm"
)

type TokensStore interface {
	Create(ctx context.Context, info token.TokenInfo) error

	RemoveByCode(ctx context.Context, code string) error

	RemoveByAccess(ctx context.Context, access string) error

	RemoveByRefresh(ctx context.Context, refresh string) error

	GetByCode(ctx context.Context, code string) (*token.TokenInfo, error)

	GetByAccess(ctx context.Context, access string) (*token.TokenInfo, error)

	GetByRefresh(ctx context.Context, refresh string) (*token.TokenInfo, error)
}

type TokenStore struct {
	db *gorm.DB
}

func NewMemoryTokenStore(db *gorm.DB) (*TokenStore, error) {
	return &TokenStore{db: db}, nil
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

func (ts *TokenStore) GetByCode(ctx context.Context, code string) (*token.TokenInfo, error) {
	return nil, nil
}

func (ts *TokenStore) GetByAccess(ctx context.Context, access string) (*token.TokenInfo, error) {
	return nil, nil
}

func (ts *TokenStore) GetByRefresh(ctx context.Context, refresh string) (*token.TokenInfo, error) {
	return nil, nil
}
