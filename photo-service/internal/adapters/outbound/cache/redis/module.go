package redis

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
)

var Module = fx.Module("cache.redis",
	fx.Provide(NewClient),
	fx.Provide(NewCache),
	fx.Provide(func(c *Cache) ports.CacheRepository { return c }),
)
