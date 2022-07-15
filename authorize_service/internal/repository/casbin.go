package repository

import (
	"time"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"go.uber.org/zap"
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

	query := c.db.Raw(`select * from casbin_rule;`).Scan(&casbinRules)
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
	query := c.db.Raw(`select * from casbin_rule where id = ?;`, id).Scan(&casbinRule)
	if query.Error != nil {
		return nil, query.Error
	}

	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}

	return &casbinRule, nil
}

func (c *CasbinRepo) CreateCasbinRule(casbin model.CasbinRule) error {
	stmtString := "INSERT INTO casbin_rule(id,ptype,v0,v1,v2) VALUES (?,?,?,?,?);"
	query := c.db.Raw(stmtString, casbin.ID, casbin.Ptype, casbin.V0, casbin.V1, casbin.V2)
	if query.RowsAffected == 0 {
		return common.NotFound
	}
	if query.Error != nil {
		log.Log().Error(query.Error.Error(), zap.Duration("backoff", time.Second))
		return query.Error
	}

	return nil
}

func (c *CasbinRepo) DeleteCasbinRule(id int) error {
	stmtString := "DELETE FROM `casbin_rule` where id = ?;"
	err := c.db.Raw(stmtString, id).Error
	if err != nil {
		log.Log().Error(err.Error(), zap.Duration("backoff", time.Second))
		return err
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRulePtype(id int, ptype string) error {
	stmtString := "UPDATE `casbin_rule` SET ptype = ?  WHERE id = ?;"
	err := c.db.Raw(stmtString, ptype, id).Error
	if err != nil {
		log.Log().Error(err.Error(), zap.Duration("backoff", time.Second))
		return err
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleName(id int, name string) error {
	stmtString := "UPDATE `casbin_rule` SET v0 = ?  WHERE id = ?;"
	err := c.db.Raw(stmtString, name, id).Error
	if err != nil {
		log.Log().Error(err.Error(), zap.Duration("backoff", time.Second))
		return err
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinRuleEndpoint(id int, endpoint string) error {
	stmtString := "UPDATE `casbin_rule` SET v1 = ? WHERE id = ?;"
	err := c.db.Exec(stmtString, endpoint, id).Error
	if err != nil {
		log.Log().Error(err.Error(), zap.Duration("backoff", time.Second))
		return err
	}

	return nil
}

func (c *CasbinRepo) UpdateCasbinMethod(id int, method string) error {
	stmtString := "UPDATE `casbin_rule` SET v2 = ?  WHERE id = ?;"
	err := c.db.Exec(stmtString, method, id)
	if err != nil {
		log.Log().Error(err.Error.Error(), zap.Duration("backoff", time.Second))
		return err.Error
	}

	return nil
}
