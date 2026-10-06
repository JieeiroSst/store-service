package application

import (
	"context"

	"github.com/JIeeiroSst/serpapi-service/config"
)

type Limiter struct{ slots chan struct{} }

func NewLimiter(cfg *config.Config) *Limiter {
	return &Limiter{slots: make(chan struct{}, cfg.Search.MaxConcurrent)}
}

func (l *Limiter) Acquire(ctx context.Context) (func(), error) {
	select {
	case l.slots <- struct{}{}:
		return func() { <-l.slots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
