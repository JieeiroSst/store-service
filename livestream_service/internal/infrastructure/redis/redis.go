package redis

import (
	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func NewClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

var Module = fx.Options(
	fx.Provide(NewClient),
)
