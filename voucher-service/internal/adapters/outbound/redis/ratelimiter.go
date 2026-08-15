package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, limit int) (bool, error) {
	now := time.Now().UTC()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/60)

	count, err := r.client.Incr(ctx, windowKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		r.client.Expire(ctx, windowKey, time.Minute)
	}
	return count <= int64(limit), nil
}
