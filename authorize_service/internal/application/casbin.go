package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"github.com/casbin/casbin/v2"
	casbinpersist "github.com/casbin/casbin/v2/persist"
	"go.uber.org/zap"
)

type casbinService struct {
	repo    port.CasbinRepository
	adapter casbinpersist.Adapter
	cache   port.CachePort

	mu          sync.RWMutex
	enforcer    *casbin.Enforcer
	enforcerExp time.Time
}

func NewCasbinService(
	repo port.CasbinRepository,
	adapter casbinpersist.Adapter,
	cache port.CachePort,
) port.CasbinUsecase {
	return &casbinService{
		repo:    repo,
		adapter: adapter,
		cache:   cache,
	}
}

// ─── enforcer lifecycle ───────────────────────────────────────────────────────

func (s *casbinService) getEnforcer(ctx context.Context) (*casbin.Enforcer, error) {
	lg := logger.WithContext(ctx)

	s.mu.RLock()
	if s.enforcer != nil && time.Now().Before(s.enforcerExp) {
		defer s.mu.RUnlock()
		return s.enforcer, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock.
	if s.enforcer != nil && time.Now().Before(s.enforcerExp) {
		return s.enforcer, nil
	}

	e, err := casbin.NewEnforcer(common.RBACModelPath, s.adapter)
	if err != nil {
		lg.Error("getEnforcer: NewEnforcer failed", zap.Error(err))
		return nil, common.ErrEnforcerFailed
	}
	if err := e.LoadPolicy(); err != nil {
		lg.Error("getEnforcer: LoadPolicy failed", zap.Error(err))
		return nil, common.ErrDBFailed
	}

	s.enforcer = e
	s.enforcerExp = time.Now().Add(common.CacheTTLEnforcer * time.Second)
	return e, nil
}

func (s *casbinService) invalidateEnforcer() {
	s.mu.Lock()
	s.enforcerExp = time.Time{}
	s.mu.Unlock()
}

// ─── port.CasbinUsecase implementation ───────────────────────────────────────

func (s *casbinService) Enforce(ctx context.Context, auth model.CasbinAuth) error {
	lg := logger.WithContext(ctx)

	e, err := s.getEnforcer(ctx)
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
	s.invalidateEnforcer()
	return nil
}

func (s *casbinService) DeleteRule(ctx context.Context, id int) error {
	lg := logger.WithContext(ctx)

	if err := s.repo.DeleteCasbinRule(ctx, id); err != nil {
		lg.Error("DeleteRule", zap.Error(err))
		return err
	}
	s.invalidateEnforcer()
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
	s.invalidateEnforcer()
	return nil
}
