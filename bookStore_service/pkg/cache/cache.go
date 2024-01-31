package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{
		rdb: rdb,
	}
}

func (r *Cache) HMSet(ctx context.Context, key string, object interface{}) {
	r.rdb.HMSet(ctx, key, object)
}

func (r *Cache) HMGet(ctx context.Context, key string, object interface{}) error {
	scores := r.rdb.HGetAll(ctx, key).Scan(&object)
	return scores
}
