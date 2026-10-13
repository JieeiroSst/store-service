package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(New),
)

func New(lc fx.Lifecycle, cfg *config.Config) *redis.Client {
	if !cfg.Redis.Enabled() {
		slog.Warn("REDIS_ADDR is not set, caching and rate limiting are disabled")
		return nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  cfg.Redis.OpTimeout,
		WriteTimeout: cfg.Redis.OpTimeout,
		PoolSize:     20,
		MaxRetries:   1,
	})
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				slog.Warn("redis is unreachable, continuing without cache until it recovers", "addr", cfg.Redis.Addr, "err", err)
			}
			return nil
		},
		OnStop: func(context.Context) error { return client.Close() },
	})
	return client
}
