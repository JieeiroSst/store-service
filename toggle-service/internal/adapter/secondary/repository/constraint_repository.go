package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type constraintRepository struct {
	db *gorm.DB
}

func NewConstraintRepository(db *gorm.DB) port.ConstraintRepository {
	return &constraintRepository{db: db}
}

func (r *constraintRepository) Create(ctx context.Context, c *model.Constraint) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *constraintRepository) ListByStrategy(ctx context.Context, strategyID uuid.UUID) ([]model.Constraint, error) {
	var constraints []model.Constraint
	if err := r.db.WithContext(ctx).Where("strategy_id = ?", strategyID).Find(&constraints).Error; err != nil {
		return nil, err
	}
	return constraints, nil
}

func (r *constraintRepository) Update(ctx context.Context, c *model.Constraint) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *constraintRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Constraint{}, "id = ?", id).Error
}

func (r *constraintRepository) DeleteByStrategy(ctx context.Context, strategyID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Constraint{}, "strategy_id = ?", strategyID).Error
}
