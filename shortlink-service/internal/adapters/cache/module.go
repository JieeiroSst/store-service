package cache

import (
	"context"

	"github.com/JIeeiroSst/shortlink-service/internal/config"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("cache",
	fx.Provide(NewCacheFromConfig),
)

func NewCacheFromConfig(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) ports.Cache {
	if cfg.RedisURL == "" {
		return NewNoopCache()
	}

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Warn("invalid REDIS_URL, disabling cache", zap.Error(err))
		return NewNoopCache()
	}
	client := redis.NewClient(opts)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return NewRedisCache(client, log)
}
