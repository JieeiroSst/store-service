package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/redis/go-redis/v9"
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
	casbinRepo  repository.Casbins
	snowflake   snowflake.SnowflakeData
	adapter     persist.Adapter
	cacheHelper cache.CacheHelper
}

func NewCasbinUsecase(casbinRepo repository.Casbins,
	snowflake snowflake.SnowflakeData, adapter persist.Adapter,
	cacheHelper cache.CacheHelper) *CasbinUsecase {
	return &CasbinUsecase{
		casbinRepo:  casbinRepo,
		snowflake:   snowflake,
		adapter:     adapter,
		cacheHelper: cacheHelper,
	}
}

func (a *CasbinUsecase) EnforceCasbin(auth model.CasbinAuth) error {
	enforcer, err := casbin.NewEnforcer(common.RBAC_MODEL, a.adapter)
	if err != nil {
		log.Error(common.Failedenforcer.Error())
		return common.Failedenforcer
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		log.Error(common.FailedDB.Error())
		return common.FailedDB
	}
	ok, err := enforcer.Enforce(auth.Sub, auth.Obj, auth.Act)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if !ok {
		log.Error(common.NotAllow.Error())
		return common.NotAllow
	}
	return nil
}

func (a *CasbinUsecase) CasbinRuleAll() ([]model.CasbinRule, error) {
	var (
		casbins []model.CasbinRule
		errDB   error
	)
	valueIntrface, err := a.cacheHelper.GetInterface(context.Background(), common.ListCasbinKeyCache, casbins)
	if err != nil {
		casbins, errDB = a.casbinRepo.CasbinRuleAll()
		if errDB != nil {
			log.Error(err.Error())
			return nil, err
		}
		if err == redis.Nil {
			_ = a.cacheHelper.Set(context.Background(), common.ListCasbinKeyCache, casbins, time.Second*60)
		}
	} else {
		casbins = valueIntrface.([]model.CasbinRule)
	}

	return casbins, nil
}

func (a *CasbinUsecase) CasbinRuleById(id int) (*model.CasbinRule, error) {
	var (
		casbin *model.CasbinRule
		errDB  error
	)
	valueInterface, err := a.cacheHelper.GetInterface(context.Background(), fmt.Sprintf(common.CasbinByIDKeyCache, id), casbin)
	if err != nil {
		casbin, errDB = a.casbinRepo.CasbinRuleById(id)
		if errDB != nil {
			log.Error(err.Error())
			return nil, err
		}
		if err == redis.Nil {
			_ = a.cacheHelper.Set(context.Background(), fmt.Sprintf(common.CasbinByIDKeyCache, id), casbin, time.Second*60)
		}
	} else {
		casbin = valueInterface.(*model.CasbinRule)
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
		log.Error(err.Error())
		return err
	}

	return nil
}

func (a *CasbinUsecase) DeleteCasbinRule(id int) error {
	if err := a.casbinRepo.DeleteCasbinRule(id); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRulePtype(id int, ptype string) error {
	if err := a.casbinRepo.UpdateCasbinRulePtype(id, ptype); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleName(id int, name string) error {
	if err := a.casbinRepo.UpdateCasbinRuleName(id, name); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleEndpoint(id int, endpoint string) error {
	if err := a.casbinRepo.UpdateCasbinRuleEndpoint(id, endpoint); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinMethod(id int, method string) error {
	if err := a.casbinRepo.UpdateCasbinMethod(id, method); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
