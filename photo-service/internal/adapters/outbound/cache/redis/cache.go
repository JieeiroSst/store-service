package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

const keyPrefix = "photo-service:"

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

var _ ports.CacheRepository = (*Cache)(nil)

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, keyPrefix+key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrCacheMiss
		}
		return "", err
	}
	return val, nil
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, keyPrefix+key, value, ttl).Err()
}
