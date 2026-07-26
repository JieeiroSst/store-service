package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
)

type adTargetingRuleRepository struct {
	db *gorm.DB
}

func NewAdTargetingRuleRepository(db *gorm.DB) port.AdTargetingRuleRepository {
	return &adTargetingRuleRepository{db: db}
}

func (r *adTargetingRuleRepository) Create(ctx context.Context, rule *model.AdTargetingRule) error {
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adTargetingRuleRepository) GetByID(ctx context.Context, id uint) (*model.AdTargetingRule, error) {
	var rule model.AdTargetingRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &rule, nil
}

func (r *adTargetingRuleRepository) Update(ctx context.Context, rule *model.AdTargetingRule) error {
	if err := r.db.WithContext(ctx).Model(&model.AdTargetingRule{}).
		Where("id = ?", rule.ID).
		Updates(rule).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adTargetingRuleRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.AdTargetingRule{}, id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adTargetingRuleRepository) List(ctx context.Context) ([]model.AdTargetingRule, error) {
	var rules []model.AdTargetingRule
	if err := r.db.WithContext(ctx).Order("id desc").Find(&rules).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return rules, nil
}

func (r *adTargetingRuleRepository) ListActiveByAd(ctx context.Context, adID uint) ([]model.AdTargetingRule, error) {
	var rules []model.AdTargetingRule
	if err := r.db.WithContext(ctx).
		Where("ad_id = ? AND is_active = ?", adID, true).
		Find(&rules).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return rules, nil
}
