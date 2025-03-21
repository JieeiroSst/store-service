package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	UpdateCasbinRule(ctx context.Context, id int, filed, value string) error
}

type CasbinUsecase struct {
	casbinRepo  repository.Casbins
	adapter     persist.Adapter
	cacheHelper cache.CacheHelper
}

func NewCasbinUsecase(casbinRepo repository.Casbins, adapter persist.Adapter,
	cacheHelper cache.CacheHelper) *CasbinUsecase {
	return &CasbinUsecase{
		casbinRepo:  casbinRepo,
		adapter:     adapter,
		cacheHelper: cacheHelper,
	}
}

func (a *CasbinUsecase) EnforceCasbin(ctx context.Context, auth model.CasbinAuth) error {
	lg := logger.WithContext(ctx)
	enforcer, err := casbin.NewEnforcer(common.RBAC_MODEL, a.adapter)
	if err != nil {
		lg.Error("EnforceCasbin", zap.Error(err))
		return common.Failedenforcer
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		lg.Error("LoadPolicy", zap.Error(common.FailedDB))
		return common.FailedDB
	}
	ok, err := enforcer.Enforce(auth.Sub, auth.Obj, auth.Act)
	if err != nil {
		lg.Error("Enforce", zap.Error(err))
		return err
	}
	if !ok {
		lg.Error("LoadPolicy", zap.Error(common.NotAllow))
		return common.NotAllow
	}
	return nil
}

func (a *CasbinUsecase) CasbinRuleAll(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error) {
	lg := logger.WithContext(ctx)
	casbins, err := a.casbinRepo.CasbinRuleAll(ctx, p)
	if err != nil {
		lg.Error("CasbinRuleAll", zap.Error(err))
		return pagination.Pagination{}, err
	}
	lg.Info("CasbinRuleAll", zap.Any("casbins", casbins))
	return casbins, nil
}

func (a *CasbinUsecase) CasbinRuleById(ctx context.Context, id int) (*model.CasbinRule, error) {
	lg := logger.WithContext(ctx)
	var (
		casbin *model.CasbinRule
		errDB  error
	)
	valueInterface, err := a.cacheHelper.GetInterface(context.Background(), fmt.Sprintf(common.CasbinByIDKeyCache, id), casbin)
	if err != nil {
		casbin, errDB = a.casbinRepo.CasbinRuleById(ctx, id)
		if errDB != nil {
			lg.Error("CasbinRuleById", zap.Error(err))
			return nil, err
		}
		if err == redis.Nil {
			_ = a.cacheHelper.Set(context.Background(), fmt.Sprintf(common.CasbinByIDKeyCache, id), casbin, time.Second*60)
		}
	} else {
		casbin = valueInterface.(*model.CasbinRule)
	}
	lg.Info("CasbinRuleById", zap.Any("casbin", casbin))
	return casbin, nil
}

func (a *CasbinUsecase) CreateCasbinRule(ctx context.Context, casbin model.CasbinRule) error {
	lg := logger.WithContext(ctx)
	object := model.CasbinRule{
		ID:    geared_id.GearedIntID(),
		Ptype: casbin.Ptype,
		V0:    casbin.V0,
		V1:    casbin.V1,
		V2:    casbin.V2,
	}

	if err := a.casbinRepo.CreateCasbinRule(ctx, object); err != nil {
		lg.Error("CreateCasbinRule", zap.Error(err))
		return err
	}

	return nil
}

func (a *CasbinUsecase) DeleteCasbinRule(ctx context.Context, id int) error {
	lg := logger.WithContext(ctx)
	if err := a.casbinRepo.DeleteCasbinRule(ctx, id); err != nil {
		lg.Error("DeleteCasbinRule", zap.Error(err))
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRulePtype(ctx context.Context, id int, ptype string) error {
	lg := logger.WithContext(ctx)
	if err := a.casbinRepo.UpdateCasbinRulePtype(ctx, id, ptype); err != nil {
		lg.Error("UpdateCasbinRulePtype", zap.Error(err))
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleName(ctx context.Context, id int, name string) error {
	lg := logger.WithContext(ctx)
	if err := a.casbinRepo.UpdateCasbinRuleName(ctx, id, name); err != nil {
		lg.Error("UpdateCasbinRuleName", zap.Error(err))
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRuleEndpoint(ctx context.Context, id int, endpoint string) error {
	lg := logger.WithContext(ctx)
	if err := a.casbinRepo.UpdateCasbinRuleEndpoint(ctx, id, endpoint); err != nil {
		lg.Error("UpdateCasbinRuleEndpoint", zap.Error(err))
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinMethod(ctx context.Context, id int, method string) error {
	lg := logger.WithContext(ctx)
	if err := a.casbinRepo.UpdateCasbinMethod(ctx, id, method); err != nil {
		lg.Error("UpdateCasbinMethod", zap.Error(err))
		return err
	}
	return nil
}

func (a *CasbinUsecase) UpdateCasbinRule(ctx context.Context, id int, filed, value string) error {
	lg := logger.WithContext(ctx)
	switch filed {
	case common.PTYPE:
		if err := a.UpdateCasbinRulePtype(ctx, id, value); err != nil {
			lg.Error("UpdateCasbinRule", zap.Any("UpdateCasbinRulePtype", err))
			return err
		}
	case common.NAME:
		if err := a.UpdateCasbinRuleName(ctx, id, value); err != nil {
			lg.Error("UpdateCasbinRule", zap.Any("UpdateCasbinRulePtype", err))
			return err
		}
	case common.ENDPOINT:
		if err := a.UpdateCasbinRuleEndpoint(ctx, id, value); err != nil {
			lg.Error("UpdateCasbinRule", zap.Any("UpdateCasbinRuleEndpoint", err))
			return err
		}
	case common.METHOD:
		if err := a.UpdateCasbinMethod(ctx, id, value); err != nil {
			lg.Error("UpdateCasbinRule", zap.Any("UpdateCasbinMethod", err))
			return err
		}
	}
	return nil
}
