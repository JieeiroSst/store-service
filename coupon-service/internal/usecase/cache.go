package usecase

import (
	"github.com/redis/go-redis/v9"
)

type CacheUsecase interface {
}

type cacheUsecase struct {
	redis *redis.ClusterClient
}

func NewCacheUsecase(redis *redis.ClusterClient) CacheUsecase {
	return &cacheUsecase{redis: redis}
}
