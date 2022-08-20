package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	redis *redis.Client
}

func NewRedis(host string) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &Cache{
		redis: rdb,
	}
}

func (r *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (r *Cache) Set(ctx context.Context, key string, value []byte) error {
	err := r.redis.Set(ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}
	return nil
}
