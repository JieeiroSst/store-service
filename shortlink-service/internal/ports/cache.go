package ports

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string, ttl time.Duration)
	Del(ctx context.Context, keys ...string)
	Enabled() bool
	Ping(ctx context.Context) error
}

func LinkResolutionCacheKey(shortCode, templateSlug string) string {
	if templateSlug != "" {
		return "link:" + templateSlug + ":" + shortCode
	}
	return "link:" + shortCode
}
