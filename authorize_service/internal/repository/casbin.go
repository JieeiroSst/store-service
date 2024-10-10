package repository

import (
	"context"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"gorm.io/gorm"
)

type Casbins interface {
	CasbinRuleAll(ctx context.Context) ([]model.CasbinRule, error)
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

func (c *CasbinRepo) CasbinRuleAll(ctx context.Context) ([]model.CasbinRule, error) {
	log := logger.ConfigZap()
	var casbinRules []model.CasbinRule
	query := c.db.Table("casbin_rule").Find(&casbinRules)
	if query.Error != nil {
		log.Error(query.Error.Error())
		return nil, query.Error
	}
	if query.RowsAffected == 0 {
		log.Error(common.NotFound.Error())
		return nil, common.NotFound
	}
	log.Info(casbinRules)
	return casbinRules, nil
}

func (c *CasbinRepo) CasbinRuleById(ctx context.Context, id int) (*model.CasbinRule, error) {
	var casbinRule model.CasbinRule
	log := logger.ConfigZap()
	query := c.db.Table("casbin_rule").Where("id = ?", id).Find(&casbinRule)
	if query.Error != nil {
		log.Error(query.Error.Error())
		return nil, query.Error
	}

	if query.RowsAffected == 0 {
		log.Error(common.NotFound.Error())
		return nil, common.NotFound
	}
	log.Info(casbinRule)
	return &casbinRule, nil
}

func (c *CasbinRepo) CreateCasbinRule(ctx context.Context, casbin model.CasbinRule) error {
	log := logger.ConfigZap()
	query := c.db.Table("casbin_rule").Save(&casbin)
	if query.RowsAffected == 0 {
		log.Error(common.NotFound.Error())
		return common.NotFound
	}
	if query.Error != nil {
		log.Error(query.Error.Error())
		return query.Error
	}

	return nil
}

func (c *CasbinRepo) DeleteCasbinRule(ctx context.Context, id int) error {
	log := logger.ConfigZap()
	stmtString := "DELETE FROM `casbin_rule` where id = ?;"
	query := c.db.Raw(stmtString, id)
	if query.Error != nil {
		log.Error(query.Error)
		return query.Error
	}
	if query.RowsAffected == 0 {
		log.Error(common.NotFound)
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRulePtype(ctx context.Context, id int, ptype string) error {
	log := logger.ConfigZap()
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("ptype", ptype)
	if query.Error != nil {
		log.Error(query.Error.Error())
		return query.Error
	}
	if query.RowsAffected == 0 {
		log.Error(common.NotFound.Error())
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleName(ctx context.Context, id int, name string) error {
	log := logger.ConfigZap()
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v0", name)
	if query.Error != nil {
		log.Error(query.Error.Error())
		return query.Error
	}
	if query.RowsAffected == 0 {
		log.Error(common.NotFound.Error())
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleEndpoint(ctx context.Context, id int, endpoint string) error {
	log := logger.ConfigZap()
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v1", endpoint)
	if query.Error != nil {
		log.Error(query.Error.Error())
		return query.Error
	}
	if query.RowsAffected == 0 {
		log.Error(common.NotFound.Error())
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinMethod(ctx context.Context, id int, method string) error {
	log := logger.ConfigZap()
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v2", method)
	if query.Error != nil {
		log.Error(query.Error.Error())
		return query.Error
	}
	if query.RowsAffected == 0 {
		log.Error(common.NotFound.Error())
		return common.NotFound
	}

	return nil
}
