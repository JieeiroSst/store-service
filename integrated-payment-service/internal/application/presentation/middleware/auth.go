package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/integrated-payment-service/internal/infrastructure/userservice"
	"github.com/gin-gonic/gin"
)

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (userservice.Identity, error)
}

const (
	contextKeyUserID = "userID"
	contextKeyRole   = "role"
)

func RequireAuth(authn Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if authHeader == "" || tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bearer token required"})
			c.Abort()
			return
		}

		id, err := authn.Authenticate(c.Request.Context(), tokenString)
		if err != nil {
			if errors.Is(err, userservice.ErrUpstream) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "authentication service unavailable"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			}
			c.Abort()
			return
		}

		c.Set(contextKeyUserID, id.UserID)
		c.Set(contextKeyRole, id.Role)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	v, _ := c.Get(contextKeyUserID)
	s, _ := v.(string)
	return s
}

func UserRole(c *gin.Context) string {
	v, _ := c.Get(contextKeyRole)
	s, _ := v.(string)
	return s
}

func IsAdmin(c *gin.Context) bool {
	return userservice.Identity{Role: UserRole(c)}.IsAdmin()
}
