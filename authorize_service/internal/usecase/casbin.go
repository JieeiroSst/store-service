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
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/redis/go-redis/v9"
)

type Casbins interface {
	EnforceCasbin(ctx context.Context, auth model.CasbinAuth) error
	CasbinRuleAll(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error)
	CasbinRuleById(ctx context.Context, id int) (*model.CasbinRule, error)
	CreateCasbinRule(ctx context.Context, casbin model.CasbinRule) error
	DeleteCasbinRule(ctx context.Context, id int) error
	UpdateCasbinRulePtype(ctx context.Context, id int, ptype string) error
	UpdateCasbinRuleName(ctx context.Context, id int, name string) error
	UpdateCasbinRuleEndpoint(ctx context.Context, id int, endpoint string) error
	UpdateCasbinMethod(ctx context.Context, id int, method string) error
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

func (a *CasbinUsecase) EnforceCasbin(ctx context.Context, auth model.CasbinAuth) error {
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

func (a *CasbinUsecase) CasbinRuleAll(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error) {
	casbins, err := a.casbinRepo.CasbinRuleAll(ctx, p)
	if err != nil {
		return pagination.Pagination{}, err
	}

	return casbins, nil
}

func (a *CasbinUsecase) CasbinRuleById(ctx context.Context, id int) (*model.CasbinRule, error) {
	var (
		casbin *model.CasbinRule
		errDB  error
	)
	valueInterface, err := a.cacheHelper.GetInterface(context.Background(), fmt.Sprintf(common.CasbinByIDKeyCache, id), casbin)
	if err != nil {
		casbin, errDB = a.casbinRepo.CasbinRuleById(ctx, id)
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

func (a *CasbinUsecase) CreateCasbinRule(ctx context.Context, casbin model.CasbinRule) error {
	object := model.CasbinRule{
		ID:    a.snowflake.GearedID(),
		Ptype: casbin.Ptype,
		V0:    casbin.V0,
		V1:    casbin.V1,
		V2:    casbin.V2,
	}

	if err := a.casbinRepo.CreateCasbinRule(ctx, object); err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (a *CasbinUsecase) DeleteCasbinRule(ctx context.Context, id int) error {
	if err := a.casbinRepo.DeleteCasbinRule(ctx, id); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRulePtype(ctx context.Context, id int, ptype string) error {
	if err := a.casbinRepo.UpdateCasbinRulePtype(ctx, id, ptype); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleName(ctx context.Context, id int, name string) error {
	if err := a.casbinRepo.UpdateCasbinRuleName(ctx, id, name); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleEndpoint(ctx context.Context, id int, endpoint string) error {
	if err := a.casbinRepo.UpdateCasbinRuleEndpoint(ctx, id, endpoint); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinMethod(ctx context.Context, id int, method string) error {
	if err := a.casbinRepo.UpdateCasbinMethod(ctx, id, method); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
