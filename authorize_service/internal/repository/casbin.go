package repository

import (
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"gorm.io/gorm"
)

type Casbins interface {
	CasbinRuleAll() ([]model.CasbinRule, error)
	CasbinRuleById(id int) (*model.CasbinRule, error)
	CreateCasbinRule(casbin model.CasbinRule) error
	DeleteCasbinRule(id int) error
	UpdateCasbinRulePtype(id int, ptype string) error
	UpdateCasbinRuleName(id int, name string) error
	UpdateCasbinRuleEndpoint(id int, endpoint string) error
	UpdateCasbinMethod(id int, method string) error
}

type CasbinRepo struct {
	db *gorm.DB
}

func NewCasbinRepo(db *gorm.DB) *CasbinRepo {
	return &CasbinRepo{
		db: db,
	}
}

func (c *CasbinRepo) CasbinRuleAll() ([]model.CasbinRule, error) {
	var casbinRules []model.CasbinRule
	query := c.db.Table("casbin_rule").Find(&casbinRules)
	if query.Error != nil {
		return nil, query.Error
	}
	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}
	return casbinRules, nil
}

func (c *CasbinRepo) CasbinRuleById(id int) (*model.CasbinRule, error) {
	var casbinRule model.CasbinRule
	query := c.db.Table("casbin_rule").Where("id = ?", id).Find(&casbinRule)
	if query.Error != nil {
		return nil, query.Error
	}

	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}

	return &casbinRule, nil
}

func (c *CasbinRepo) CreateCasbinRule(casbin model.CasbinRule) error {
	query := c.db.Table("casbin_rule").Save(&casbin)
	if query.RowsAffected == 0 {
		return common.NotFound
	}
	if query.Error != nil {
		return query.Error
	}

	return nil
}

func (c *CasbinRepo) DeleteCasbinRule(id int) error {
	stmtString := "DELETE FROM `casbin_rule` where id = ?;"
	query := c.db.Raw(stmtString, id)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRulePtype(id int, ptype string) error {
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("ptype", ptype)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleName(id int, name string) error {
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v0", name)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleEndpoint(id int, endpoint string) error {
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v1", endpoint)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return common.NotFound
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinMethod(id int, method string) error {
	query := c.db.Table("casin_rule").Where("id = ?", id).Update("v2", method)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return common.NotFound
	}

	return nil
}
