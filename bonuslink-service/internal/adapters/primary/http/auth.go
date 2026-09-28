package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/bonuslink-service/internal/adapters/secondary/userservice"
)

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (userservice.Identity, error)
}

const ctxIdentity = "identity"

func requireAuth(authn Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || token == "" {
			errorJSON(c, http.StatusUnauthorized, "UNAUTHENTICATED", "bearer token required")
			c.Abort()
			return
		}
		id, err := authn.Authenticate(c.Request.Context(), token)
		if errors.Is(err, userservice.ErrUpstream) {
			errorJSON(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "authentication service unavailable")
			c.Abort()
			return
		}
		if err != nil {
			errorJSON(c, http.StatusUnauthorized, "UNAUTHENTICATED", "invalid token")
			c.Abort()
			return
		}
		c.Set(ctxIdentity, id)
		c.Next()
	}
}

func identity(c *gin.Context) userservice.Identity {
	id, _ := c.Get(ctxIdentity)
	v, _ := id.(userservice.Identity)
	return v
}

// isStaff: operators and admins manage rewards for every user.
func isStaff(id userservice.Identity) bool {
	return id.IsAdmin() || id.Role == "operator"
}

func requireStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isStaff(identity(c)) {
			errorJSON(c, http.StatusForbidden, "FORBIDDEN", "operator or admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}

func requireSelfOrStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := identity(c)
		if id.UserID != c.Param("user_id") && !isStaff(id) {
			errorJSON(c, http.StatusForbidden, "FORBIDDEN", "cannot read another user's rewards")
			c.Abort()
			return
		}
		c.Next()
	}
}
