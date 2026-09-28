package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/post-service/internal/adapter/secondary/userservice"
	"github.com/gin-gonic/gin"
)

// Authenticator resolves a bearer token to a user through user-service.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (userservice.Identity, error)
}

const (
	contextKeyUserID = "userID"
	contextKeyRole   = "role"
)

// RequireAuth accepts only tokens that user-service confirms are live
// sessions, so a logged-out token stops working here too.
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
