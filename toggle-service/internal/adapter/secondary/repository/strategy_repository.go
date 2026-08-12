package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type strategyRepository struct {
	db *gorm.DB
}

func NewStrategyRepository(db *gorm.DB) port.StrategyRepository {
	return &strategyRepository{db: db}
}

func (r *strategyRepository) Create(ctx context.Context, s *model.ActivationStrategy) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *strategyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ActivationStrategy, error) {
	var s model.ActivationStrategy
	if err := r.db.WithContext(ctx).Preload("Constraints").First(&s, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *strategyRepository) ListByFlagEnvironment(ctx context.Context, flagEnvironmentID uuid.UUID) ([]model.ActivationStrategy, error) {
	var strategies []model.ActivationStrategy
	if err := r.db.WithContext(ctx).
		Preload("Constraints").
		Where("feature_flag_environment_id = ?", flagEnvironmentID).
		Order("sort_order asc").
		Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

func (r *strategyRepository) Update(ctx context.Context, s *model.ActivationStrategy) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *strategyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ActivationStrategy{}, "id = ?", id).Error
}
