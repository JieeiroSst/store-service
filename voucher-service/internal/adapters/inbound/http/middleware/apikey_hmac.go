package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"

	partnerapp "github.com/JIeeiroSst/voucher-service/internal/application/partner"
	"github.com/gin-gonic/gin"
)

const partnerKeyContextKey = "partner_api_key"

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int) (bool, error)
}

func PartnerAuth(svc partnerapp.PartnerService, limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		prefix := c.GetHeader("X-API-Key-Prefix")
		timestamp := c.GetHeader("X-Timestamp")
		signature := c.GetHeader("X-Signature")
		if prefix == "" || timestamp == "" || signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "missing_signature", "message": "X-API-Key-Prefix, X-Timestamp and X-Signature headers are required"}})
			return
		}

		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		key, err := svc.VerifySignature(c.Request.Context(), prefix, timestamp, bodyBytes, signature)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "invalid_signature", "message": "signature verification failed"}})
			return
		}

		allowed, err := limiter.Allow(c.Request.Context(), key.KeyPrefix, key.RateLimitPerMin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal_error", "message": "rate limit check failed"}})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "rate_limited", "message": "too many requests"}})
			return
		}

		c.Set(partnerKeyContextKey, key)
		c.Next()
	}
}

func PartnerKeyFromGin(c *gin.Context) (*partnerapp.APIKey, bool) {
	v, ok := c.Get(partnerKeyContextKey)
	if !ok {
		return nil, false
	}
	key, ok := v.(*partnerapp.APIKey)
	return key, ok
}
