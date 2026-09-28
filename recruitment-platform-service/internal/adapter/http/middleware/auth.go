package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/userservice"
)

// Authenticator resolves a bearer token to a user through user-service,
// which issues every token; this service keeps no credentials.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (userservice.Identity, error)
}

// UserServiceAuth accepts only tokens user-service confirms are live
// sessions. It sets "user_id" (int64 user-service id) and "role" (primary
// role from authorize-service).
func UserServiceAuth(authn Authenticator, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		id, err := authn.Authenticate(c.Request.Context(), parts[1])
		if errors.Is(err, userservice.ErrUpstream) {
			logger.Warn("user-service unavailable", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "authentication service unavailable"})
			return
		}
		userID, convErr := strconv.ParseInt(id.UserID, 10, 64)
		if err != nil || convErr != nil {
			logger.Debug("invalid token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", userID)
		c.Set("role", id.Role)
		c.Next()
	}
}

// RequireRole enforces a minimum role level.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if !allowed[role.(string)] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
