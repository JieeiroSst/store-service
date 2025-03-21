package repository

import (
	"context"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Casbins interface {
	CasbinRuleAll(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error)
	CasbinRuleById(ctx context.Context, id int) (*model.CasbinRule, error)
	CreateCasbinRule(ctx context.Context, casbin model.CasbinRule) error
	DeleteCasbinRule(ctx context.Context, id int) error
	UpdateCasbinRulePtype(ctx context.Context, id int, ptype string) error
	UpdateCasbinRuleName(ctx context.Context, id int, name string) error
	UpdateCasbinRuleEndpoint(ctx context.Context, id int, endpoint string) error
	UpdateCasbinMethod(ctx context.Context, id int, method string) error
}

type CasbinRepo struct {
	db *gorm.DB
}

func NewCasbinRepo(db *gorm.DB) *CasbinRepo {
	return &CasbinRepo{
		db: db,
	}
}

func (c *CasbinRepo) CasbinRuleAll(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error) {
	lg := logger.WithContext(ctx)
	var casbinRules []model.CasbinRule

	c.db.Scopes(pagination.Paginate(casbinRules, &p, c.db)).Find(&casbinRules)
	p.Rows = casbinRules

	lg.Info("CasbinRuleAll", zap.Any("casbinRules", casbinRules))

	return p, nil
}

func (c *CasbinRepo) CasbinRuleById(ctx context.Context, id int) (*model.CasbinRule, error) {
	lg := logger.WithContext(ctx)
	var casbinRule model.CasbinRule
	query := c.db.Table("casbin_rule").Where("id = ?", id).Find(&casbinRule)
	if query.Error != nil {
		lg.Error("CasbinRuleById", zap.Error(query.Error))
		return nil, query.Error
	}

	if query.RowsAffected == 0 {
		lg.Error("CasbinRuleById", zap.Error(common.NotFound))
		return nil, common.NotFound
	}
	lg.Info("CasbinRuleById", zap.Any("casbinRule", casbinRule))
	return &casbinRule, nil
}

func (c *CasbinRepo) CreateCasbinRule(ctx context.Context, casbin model.CasbinRule) error {
	lg := logger.WithContext(ctx)
	query := c.db.Table("casbin_rule").Save(&casbin)
	if query.RowsAffected == 0 {
		lg.Error("CreateCasbinRule", zap.Error(common.NotFound))
		return common.NotFound
	}
	if query.Error != nil {
		lg.Error("CreateCasbinRule", zap.Error(query.Error))
		return query.Error
	}

	return nil
}

func (c *CasbinRepo) DeleteCasbinRule(ctx context.Context, id int) error {
	lg := logger.WithContext(ctx)
	stmtString := "DELETE FROM `casbin_rule` where id = ?;"
	query := c.db.Raw(stmtString, id)
	if query.Error != nil {
		lg.Error("DeleteCasbinRule", zap.Error(query.Error))
		return query.Error
	}
	if query.RowsAffected == 0 {
		lg.Error("DeleteCasbinRule", zap.Error(common.NotFound))
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRulePtype(ctx context.Context, id int, ptype string) error {
	lg := logger.WithContext(ctx)
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("ptype", ptype)
	if query.Error != nil {
		lg.Error("UpdateCasbinRulePtype", zap.Error(query.Error))
		return query.Error
	}
	if query.RowsAffected == 0 {
		lg.Error("UpdateCasbinRulePtype", zap.Error(common.NotFound))
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleName(ctx context.Context, id int, name string) error {
	lg := logger.WithContext(ctx)
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v0", name)
	if query.Error != nil {
		lg.Error("UpdateCasbinRuleName", zap.Error(query.Error))
		return query.Error
	}
	if query.RowsAffected == 0 {
		lg.Error("UpdateCasbinRuleName", zap.Error(common.NotFound))
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleEndpoint(ctx context.Context, id int, endpoint string) error {
	lg := logger.WithContext(ctx)
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v1", endpoint)
	if query.Error != nil {
		lg.Error("UpdateCasbinRuleEndpoint", zap.Error(query.Error))
		return query.Error
	}
	if query.RowsAffected == 0 {
		lg.Error("UpdateCasbinRuleEndpoint", zap.Error(common.NotFound))
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinMethod(ctx context.Context, id int, method string) error {
	lg := logger.WithContext(ctx)
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v2", method)
	if query.Error != nil {
		lg.Error("UpdateCasbinMethod", zap.Error(query.Error))
		return query.Error
	}
	if query.RowsAffected == 0 {
		lg.Error("UpdateCasbinMethod", zap.Error(common.NotFound))
		return common.NotFound
	}

	return nil
}
