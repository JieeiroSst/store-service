package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"go.uber.org/zap"
)

type casbinService struct {
	repo      port.CasbinRepository
	enforcers *EnforcerProvider
	cache     port.CachePort
}

func NewCasbinService(
	repo port.CasbinRepository,
	enforcers *EnforcerProvider,
	cache port.CachePort,
) port.CasbinUsecase {
	return &casbinService{
		repo:      repo,
		enforcers: enforcers,
		cache:     cache,
	}
}

// ─── port.CasbinUsecase implementation ───────────────────────────────────────

func (s *casbinService) Enforce(ctx context.Context, auth model.CasbinAuth) error {
	lg := logger.WithContext(ctx)

	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return err
	}

	ok, err := e.Enforce(auth.Sub, auth.Obj, auth.Act)
	if err != nil {
		lg.Error("Enforce", zap.Error(err))
		return err
	}
	if !ok {
		lg.Info("Enforce: access denied", zap.Any("auth", auth))
		return common.ErrNotAllowed
	}
	return nil
}

func (s *casbinService) ListRules(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error) {
	return s.repo.CasbinRuleAll(ctx, p)
}

func (s *casbinService) GetRule(ctx context.Context, id int) (*model.CasbinRule, error) {
	lg := logger.WithContext(ctx)

	cacheKey := fmt.Sprintf(common.CacheKeyCasbinByID, id)
	var rule model.CasbinRule
	if cached, err := s.cache.GetInterface(ctx, cacheKey, &rule); err == nil {
		if r, ok := cached.(*model.CasbinRule); ok {
			return r, nil
		}
	}

	r, err := s.repo.CasbinRuleByID(ctx, id)
	if err != nil {
		lg.Error("GetRule: repo", zap.Error(err))
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, r, time.Duration(common.CacheTTLCasbinByID)*time.Second)
	return r, nil
}

func (s *casbinService) CreateRule(ctx context.Context, rule model.CasbinRule) error {
	lg := logger.WithContext(ctx)

	rule.ID = geared_id.GearedIntID()
	if err := s.repo.CreateCasbinRule(ctx, rule); err != nil {
		lg.Error("CreateRule", zap.Error(err))
		return err
	}
	s.enforcers.Invalidate()
	return nil
}

func (s *casbinService) DeleteRule(ctx context.Context, id int) error {
	lg := logger.WithContext(ctx)

	if err := s.repo.DeleteCasbinRule(ctx, id); err != nil {
		lg.Error("DeleteRule", zap.Error(err))
		return err
	}
	s.enforcers.Invalidate()
	return nil
}

func (s *casbinService) UpdateRuleField(ctx context.Context, id int, field model.UpdateField, value string) error {
	lg := logger.WithContext(ctx)

	if !field.IsValid() {
		return fmt.Errorf("%w: %q", common.ErrInvalidField, field)
	}

	if err := s.repo.UpdateCasbinRuleField(ctx, id, field, value); err != nil {
		lg.Error("UpdateRuleField", zap.Error(err))
		return err
	}
	s.enforcers.Invalidate()
	return nil
}
