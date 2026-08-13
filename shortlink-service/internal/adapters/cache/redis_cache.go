package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	client *redis.Client
	log    *zap.Logger
}

func NewRedisCache(client *redis.Client, log *zap.Logger) *RedisCache {
	return &RedisCache{client: client, log: log}
}

func (c *RedisCache) Enabled() bool { return true }

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, bool) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			c.log.Warn("redis cache lookup failed, falling back to database", zap.Error(err), zap.String("key", key))
		}
		return "", false
	}
	return val, true
}

func (c *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		c.log.Warn("redis cache set failed", zap.Error(err), zap.String("key", key))
	}
}

func (c *RedisCache) Del(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	_ = c.client.Del(ctx, keys...).Err()
}

type NoopCache struct{}

func NewNoopCache() *NoopCache { return &NoopCache{} }

func (NoopCache) Enabled() bool                                      { return false }
func (NoopCache) Get(context.Context, string) (string, bool)         { return "", false }
func (NoopCache) Set(context.Context, string, string, time.Duration) {}
func (NoopCache) Del(context.Context, ...string)                     {}
func (NoopCache) Ping(context.Context) error                         { return nil }
