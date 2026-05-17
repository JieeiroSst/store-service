package cache

import (
	"context"
	"time"

	pkgcache "github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
)

type cacheAdapter struct {
	helper pkgcache.CacheHelper
}

func NewCacheAdapter(dns string) port.CachePort {
	return &cacheAdapter{helper: pkgcache.NewCacheHelper(dns)}
}

func (a *cacheAdapter) GetInterface(ctx context.Context, key string, value interface{}) (interface{}, error) {
	return a.helper.GetInterface(ctx, key, value)
}

func (a *cacheAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.helper.Set(ctx, key, value, expiration)
}

func (a *cacheAdapter) Delete(ctx context.Context, key string) error {
	return a.helper.Delete(ctx, key)
}

func (a *cacheAdapter) GetInt(ctx context.Context, key string) (int, error) {
	return a.helper.GetInt(ctx, key)
}

func (a *cacheAdapter) SetInt(ctx context.Context, key string, value int) error {
	return a.helper.SetInt(ctx, key, value)
}
