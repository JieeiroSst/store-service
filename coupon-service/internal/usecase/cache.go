package usecase

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

type CacheUsecase interface {
	SetCache(ctx context.Context, key string, value interface{}) error
	GetCache(ctx context.Context, key string) ([]byte, error)
	DelCache(ctx context.Context, key string) error
}

type cacheUsecase struct {
	redis *redis.ClusterClient
}

var (
	instance CacheUsecase
	once     sync.Once
)

func NewCacheUsecase(urls []string) CacheUsecase {
	redis := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: urls,
	})
	once.Do(func() {
		instance = &cacheUsecase{redis: redis}
	})
	return instance
}

func (c *cacheUsecase) SetCache(ctx context.Context, key string, value interface{}) error {
	return c.redis.Set(ctx, key, value, 0).Err()
}

func (c *cacheUsecase) GetCache(ctx context.Context, key string) ([]byte, error) {
	return c.redis.Get(ctx, key).Bytes()
}

func (c *cacheUsecase) DelCache(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}
