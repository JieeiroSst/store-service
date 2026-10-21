package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/wallet-service/pkg/algorithms/ratelimit"
)

const limiterIdleTTL = 10 * time.Minute

type limiterEntry struct {
	bucket   *ratelimit.TokenBucket
	lastSeen time.Time
}

func rateLimiter(capacity, ratePerSec float64) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*limiterEntry)
	lastSweep := time.Now()

	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()

		mu.Lock()
		if now.Sub(lastSweep) > limiterIdleTTL {
			for k, e := range buckets {
				if now.Sub(e.lastSeen) > limiterIdleTTL {
					delete(buckets, k)
				}
			}
			lastSweep = now
		}
		entry, ok := buckets[key]
		if !ok {
			entry = &limiterEntry{bucket: ratelimit.NewTokenBucket(capacity, ratePerSec)}
			buckets[key] = entry
		}
		entry.lastSeen = now
		mu.Unlock()

		if !entry.bucket.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errorResponse{Error: "rate limit exceeded", Code: "RATE_LIMITED"})
			return
		}
		c.Next()
	}
}
