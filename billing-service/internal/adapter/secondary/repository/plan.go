package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/billing-service/common"
	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
	"gorm.io/gorm"
)

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) port.PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(ctx context.Context, plan *model.Plan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *planRepository) GetByID(ctx context.Context, id int) (*model.Plan, error) {
	var plan model.Plan
	if err := r.db.WithContext(ctx).Where("plan_id = ?", id).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &plan, nil
}

func (r *planRepository) Update(ctx context.Context, plan *model.Plan) error {
	if err := r.db.WithContext(ctx).Model(&model.Plan{}).
		Where("plan_id = ?", plan.PlanID).
		Updates(plan).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *planRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("plan_id = ?", id).Delete(&model.Plan{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *planRepository) List(ctx context.Context) ([]model.Plan, error) {
	var plans []model.Plan
	if err := r.db.WithContext(ctx).Find(&plans).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return plans, nil
}
