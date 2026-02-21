package database

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/JIeeiroSst/wallet-service/internal/infrastructure/config"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
