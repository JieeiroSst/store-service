package cache

import (
	"context"
	"fmt"

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

func (r *Cache) Set(ctx context.Context, key string, object interface{}) {
	r.rdb.Set(ctx, key, object, 0).Err()
}

func (r *Cache) Get(ctx context.Context, key string, object interface{}) {
	err := r.rdb.Get(ctx, key).Scan(&object)
	if err != nil {
		if err == redis.Nil {
			fmt.Println("Key does not exist")
		} else {
			fmt.Println("data does not exist")
		}
	}
}
