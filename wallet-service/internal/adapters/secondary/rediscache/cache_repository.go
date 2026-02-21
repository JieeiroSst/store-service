package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

type cacheRepository struct {
	client *redis.Client
	logger *zap.Logger
}

func NewCacheRepository(client *redis.Client, logger *zap.Logger) ports.CacheRepository {
	return &cacheRepository{client: client, logger: logger}
}

func (r *cacheRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil 
	}
	if err != nil {
		r.logger.Warn("redis GET error", zap.String("key", key), zap.Error(err))
		return "", err
	}
	return val, nil
}

func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		s = string(b)
	}
	if err := r.client.Set(ctx, key, s, ttl).Err(); err != nil {
		r.logger.Warn("redis SET error", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

func (r *cacheRepository) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		r.logger.Warn("redis DEL error", zap.Strings("keys", keys), zap.Error(err))
		return err
	}
	r.logger.Debug("cache keys invalidated", zap.Strings("keys", keys))
	return nil
}

func (r *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (r *cacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *cacheRepository) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return false, err
		}
		s = string(b)
	}
	return r.client.SetNX(ctx, key, s, ttl).Result()
}
