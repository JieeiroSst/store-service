package cache

import "github.com/go-redis/redis/v8"

type CacheHelper struct {
	resdis *redis.Client
}

func NewCacheHelper(dns string) CacheHelper {
	rdb := redis.NewClient(&redis.Options{
		Addr:     dns,
		Password: "",
		DB:       0,
	})
	return CacheHelper{
		resdis: rdb,
	}
}
