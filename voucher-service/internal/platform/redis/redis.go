package redis

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewClient(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		DB:   cfg.RedisDB,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return client.Ping(ctx).Err()
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing redis client")
			return client.Close()
		},
	})

	return client, nil
}
