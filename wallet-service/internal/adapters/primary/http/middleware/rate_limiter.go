package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/wallet-service/pkg/algorithms/ratelimit"
)

type clientLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*ratelimit.TokenBucket
	capacity float64
	rate     float64
}

func RateLimiter(capacity, ratePerSec float64) gin.HandlerFunc {
	cl := &clientLimiter{
		buckets:  make(map[string]*ratelimit.TokenBucket),
		capacity: capacity,
		rate:     ratePerSec,
	}
	return func(c *gin.Context) {
		clientID := c.ClientIP()
		cl.mu.RLock()
		bucket, ok := cl.buckets[clientID]
		cl.mu.RUnlock()

		if !ok {
			cl.mu.Lock()
			bucket = ratelimit.NewTokenBucket(cl.capacity, cl.rate)
			cl.buckets[clientID] = bucket
			cl.mu.Unlock()
		}

		if !bucket.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"code":  "RATE_LIMITED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func IdempotencyKeyRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Idempotency-Key") == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Idempotency-Key header is required for this endpoint",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
