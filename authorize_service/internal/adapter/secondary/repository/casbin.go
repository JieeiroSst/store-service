package repository

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type casbinRepo struct {
	db *gorm.DB
}

func NewCasbinRepository(db *gorm.DB) port.CasbinRepository {
	return &casbinRepo{db: db}
}

func (r *casbinRepo) CasbinRuleAll(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error) {
	lg := logger.WithContext(ctx)

	var rules []model.CasbinRule
	r.db.Scopes(pagination.Paginate(rules, &p, r.db)).Find(&rules)
	p.Rows = rules

	lg.Info("CasbinRuleAll", zap.Int("count", len(rules)))
	return p, nil
}

func (r *casbinRepo) CasbinRuleByID(ctx context.Context, id int) (*model.CasbinRule, error) {
	lg := logger.WithContext(ctx)

	var rule model.CasbinRule
	result := r.db.Where("id = ?", id).First(&rule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		lg.Error("CasbinRuleByID", zap.Error(result.Error))
		return nil, common.ErrDBFailed
	}
	return &rule, nil
}

func (r *casbinRepo) CreateCasbinRule(ctx context.Context, rule model.CasbinRule) error {
	lg := logger.WithContext(ctx)

	if err := r.db.Create(&rule).Error; err != nil {
		lg.Error("CreateCasbinRule", zap.Error(err))
		return common.ErrDBFailed
	}
	return nil
}

func (r *casbinRepo) DeleteCasbinRule(ctx context.Context, id int) error {
	lg := logger.WithContext(ctx)

	result := r.db.Delete(&model.CasbinRule{}, id)
	if result.Error != nil {
		lg.Error("DeleteCasbinRule", zap.Error(result.Error))
		return common.ErrDBFailed
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func (r *casbinRepo) UpdateCasbinRuleField(ctx context.Context, id int, field model.UpdateField, value string) error {
	lg := logger.WithContext(ctx)

	if !field.IsValid() {
		return fmt.Errorf("%w: %q", common.ErrInvalidField, field)
	}

	result := r.db.Model(&model.CasbinRule{}).Where("id = ?", id).Update(string(field), value)
	if result.Error != nil {
		lg.Error("UpdateCasbinRuleField", zap.Error(result.Error))
		return common.ErrDBFailed
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}
