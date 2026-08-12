package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) port.TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, t *model.APIToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *tokenRepository) GetByHash(ctx context.Context, hash string) (*model.APIToken, error) {
	var t model.APIToken
	if err := r.db.WithContext(ctx).First(&t, "token_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *tokenRepository) List(ctx context.Context) ([]model.APIToken, error) {
	var tokens []model.APIToken
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *tokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.APIToken{}, "id = ?", id).Error
}
