package cache

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/port"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(New),
)

func New(lc fx.Lifecycle, cfg *config.Config) (port.Cache, error) {
	switch cfg.Cache.Backend {
	case "memory":
		return NewMemory(cfg.Cache.MaxEntries), nil
	case "none":
		return None{}, nil
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Cache.RedisAddr,
			Password: cfg.Cache.RedisPassword,
			DB:       cfg.Cache.RedisDB,
		})
		lc.Append(fx.Hook{
			OnStop: func(context.Context) error { return client.Close() },
		})
		return NewRedis(client, cfg.Cache.RedisPrefix), nil
	}
	return nil, fmt.Errorf("unknown CACHE_BACKEND %q (memory, redis or none)", cfg.Cache.Backend)
}
