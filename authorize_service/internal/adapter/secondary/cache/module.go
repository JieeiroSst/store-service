package cache

import (
	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"go.uber.org/fx"
)

func newCachePortFromConfig(cfg *config.Config) port.CachePort {
	return NewCacheAdapter(cfg.Cache.Host)
}

// Module registers the cache adapter with fx.
var Module = fx.Options(
	fx.Provide(newCachePortFromConfig),
)
