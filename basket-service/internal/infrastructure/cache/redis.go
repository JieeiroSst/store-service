package cache

import (
	"github.com/JIeeiroSst/basket-service/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// NewRedisClient does not dial eagerly (go-redis connects lazily on first
// command), so a Redis outage at startup does not fail the whole service.
func NewRedisClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Host,
		Password: cfg.Cache.Password,
	})
}

var Module = fx.Options(
	fx.Provide(NewRedisClient),
)
