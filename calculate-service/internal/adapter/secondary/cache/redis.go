package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/redis/go-redis/v9"
)

const radarCacheKey = "weather:radar:latest"

type redisCache struct {
	rdb *redis.Client
}

// NewRedisClient builds the shared Redis connection used by every cache adapter.
func NewRedisClient(dns string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: dns})
}

func NewRedisCache(rdb *redis.Client) port.WeatherCache {
	return &redisCache{rdb: rdb}
}

func NewRedisMarketCache(rdb *redis.Client) port.MarketCache {
	return &redisCache{rdb: rdb}
}

func getJSON[T any](ctx context.Context, rdb *redis.Client, key string) (*T, bool) {
	raw, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Error(ctx, "cache get %s error %v", key, err)
		}
		return nil, false
	}

	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		logger.Error(ctx, "cache unmarshal %s error %v", key, err)
		return nil, false
	}
	return &v, true
}

func setJSON(ctx context.Context, rdb *redis.Client, key string, v any, ttl time.Duration) {
	raw, err := json.Marshal(v)
	if err != nil {
		logger.Error(ctx, "cache marshal %s error %v", key, err)
		return
	}
	if err := rdb.Set(ctx, key, raw, ttl).Err(); err != nil {
		logger.Error(ctx, "cache set %s error %v", key, err)
	}
}

func (c *redisCache) GetForecast(ctx context.Context, key string) (*model.Forecast, bool) {
	return getJSON[model.Forecast](ctx, c.rdb, key)
}

func (c *redisCache) SetForecast(ctx context.Context, key string, v *model.Forecast, ttl time.Duration) {
	setJSON(ctx, c.rdb, key, v, ttl)
}

func (c *redisCache) GetTide(ctx context.Context, key string) (*model.TidePrediction, bool) {
	return getJSON[model.TidePrediction](ctx, c.rdb, key)
}

func (c *redisCache) SetTide(ctx context.Context, key string, v *model.TidePrediction, ttl time.Duration) {
	setJSON(ctx, c.rdb, key, v, ttl)
}

func (c *redisCache) GetRadar(ctx context.Context) (*model.RainRadar, bool) {
	return getJSON[model.RainRadar](ctx, c.rdb, radarCacheKey)
}

func (c *redisCache) SetRadar(ctx context.Context, v *model.RainRadar, ttl time.Duration) {
	setJSON(ctx, c.rdb, radarCacheKey, v, ttl)
}

func (c *redisCache) GetSnapshot(ctx context.Context, key string) (*model.WeatherSnapshot, bool) {
	return getJSON[model.WeatherSnapshot](ctx, c.rdb, key)
}

func (c *redisCache) SetSnapshot(ctx context.Context, key string, v *model.WeatherSnapshot, ttl time.Duration) {
	setJSON(ctx, c.rdb, key, v, ttl)
}

func (c *redisCache) GetMarketSnapshot(ctx context.Context, key string) (*model.MarketSnapshot, bool) {
	return getJSON[model.MarketSnapshot](ctx, c.rdb, key)
}

func (c *redisCache) SetMarketSnapshot(ctx context.Context, key string, v *model.MarketSnapshot, ttl time.Duration) {
	setJSON(ctx, c.rdb, key, v, ttl)
}
