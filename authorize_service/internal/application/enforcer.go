package application

import (
	"context"
	"sync"
	"time"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/casbin/casbin/v2"
	casbinpersist "github.com/casbin/casbin/v2/persist"
	"go.uber.org/zap"
)

type EnforcerProvider struct {
	adapter casbinpersist.Adapter

	mu       sync.RWMutex
	enforcer *casbin.SyncedEnforcer
	expires  time.Time
}

func NewEnforcerProvider(adapter casbinpersist.Adapter) *EnforcerProvider {
	return &EnforcerProvider{adapter: adapter}
}

func (p *EnforcerProvider) Get(ctx context.Context) (*casbin.SyncedEnforcer, error) {
	lg := logger.WithContext(ctx)

	p.mu.RLock()
	if p.enforcer != nil && time.Now().Before(p.expires) {
		defer p.mu.RUnlock()
		return p.enforcer, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.enforcer != nil && time.Now().Before(p.expires) {
		return p.enforcer, nil
	}

	if p.enforcer == nil {
		e, err := casbin.NewSyncedEnforcer(common.RBACModelPath, p.adapter)
		if err != nil {
			lg.Error("EnforcerProvider: NewSyncedEnforcer failed", zap.Error(err))
			return nil, common.ErrEnforcerFailed
		}
		p.enforcer = e
	}
	if err := p.enforcer.LoadPolicy(); err != nil {
		lg.Error("EnforcerProvider: LoadPolicy failed", zap.Error(err))
		return nil, common.ErrDBFailed
	}

	p.expires = time.Now().Add(common.CacheTTLEnforcer * time.Second)
	return p.enforcer, nil
}

func (p *EnforcerProvider) Invalidate() {
	p.mu.Lock()
	p.expires = time.Time{}
	p.mu.Unlock()
}
