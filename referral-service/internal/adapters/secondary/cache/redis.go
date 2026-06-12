package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/ports"
)

var Module = fx.Options(
	fx.Provide(NewRedisCache),
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(cfg *config.Config) (ports.Cache, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Endpoint, cfg.Redis.Port)
	opts := &redis.Options{
		Addr:         addr,
		Username:     cfg.Redis.Username,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinConnection,
		DialTimeout:  cfg.Redis.Timeout,
		ReadTimeout:  cfg.Redis.Timeout,
		WriteTimeout: cfg.Redis.Timeout,
	}

	if cfg.Redis.TLS == "true" {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &redisCache{client: client}, nil
}

func (c *redisCache) GetInt64(ctx context.Context, key string) (int64, bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return n, true, nil
}

func (c *redisCache) SetInt64(ctx context.Context, key string, value int64, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Incr atomically increments the counter. If the key did not exist (result == 1),
// it sets the TTL so the key expires at the intended time.
func (c *redisCache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	n, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		c.client.Expire(ctx, key, ttl) //nolint:errcheck
	}
	return n, nil
}
