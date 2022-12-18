package usecase

import (
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
)

type Casbins interface {
	EnforceCasbin(auth model.CasbinAuth) error
	CasbinRuleAll() ([]model.CasbinRule, error)
	CasbinRuleById(id int) (*model.CasbinRule, error)
	CreateCasbinRule(casbin model.CasbinRule) error
	DeleteCasbinRule(id int) error
	UpdateCasbinRulePtype(id int, ptype string) error
	UpdateCasbinRuleName(id int, name string) error
	UpdateCasbinRuleEndpoint(id int, endpoint string) error
	UpdateCasbinMethod(id int, method string) error
}

type CasbinUsecase struct {
	casbinRepo repository.Casbins
	snowflake  snowflake.SnowflakeData
	adapter    persist.Adapter
}

func NewCasbinUsecase(casbinRepo repository.Casbins,
	snowflake snowflake.SnowflakeData, adapter persist.Adapter) *CasbinUsecase {
	return &CasbinUsecase{
		casbinRepo: casbinRepo,
		snowflake:  snowflake,
		adapter:    adapter,
	}
}

func (a *CasbinUsecase) EnforceCasbin(auth model.CasbinAuth) error {
	enforcer, err := casbin.NewEnforcer("config/conf/rbac_model.conf", a.adapter)
	if err != nil {
		log.Error(common.Failedenforcer)
		return common.Failedenforcer
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		log.Error(common.FailedDB)
		return common.FailedDB
	}
	ok, err := enforcer.Enforce(auth.Sub, auth.Obj, auth.Act)
	if err != nil {
		log.Error(err)
		return err
	}
	if !ok {
		log.Error(common.NotAllow)
		return common.NotAllow
	}
	return nil
}

func (a *CasbinUsecase) CasbinRuleAll() ([]model.CasbinRule, error) {
	casbins, err := a.casbinRepo.CasbinRuleAll()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return casbins, nil
}

func (a *CasbinUsecase) CasbinRuleById(id int) (*model.CasbinRule, error) {
	casbin, err := a.casbinRepo.CasbinRuleById(id)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return casbin, nil
}

func (a *CasbinUsecase) CreateCasbinRule(casbin model.CasbinRule) error {
	object := model.CasbinRule{
		ID:    a.snowflake.GearedID(),
		Ptype: casbin.Ptype,
		V0:    casbin.V0,
		V1:    casbin.V1,
		V2:    casbin.V2,
	}

	if err := a.casbinRepo.CreateCasbinRule(object); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (a *CasbinUsecase) DeleteCasbinRule(id int) error {
	if err := a.casbinRepo.DeleteCasbinRule(id); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRulePtype(id int, ptype string) error {
	if err := a.casbinRepo.UpdateCasbinRulePtype(id, ptype); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleName(id int, name string) error {
	if err := a.casbinRepo.UpdateCasbinRuleName(id, name); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleEndpoint(id int, endpoint string) error {
	if err := a.casbinRepo.UpdateCasbinRuleEndpoint(id, endpoint); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinMethod(id int, method string) error {
	if err := a.casbinRepo.UpdateCasbinMethod(id, method); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
