package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/userservice"
	"github.com/gin-gonic/gin"
)

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (userservice.Identity, error)
}

var rolePermissions = map[models.UserRole][]string{
	models.RoleAdmin:   {"*"},
	models.RoleAuditor: {"key:list", "key:read", "audit:read"},
	models.RoleUser:    {"key:create", "key:list", "key:read", "key:use", "key:rotate", "key:delete", "audit:read"},
}

func kmsRole(id userservice.Identity) models.UserRole {
	switch {
	case id.IsAdmin():
		return models.RoleAdmin
	case id.Role == string(models.RoleAuditor):
		return models.RoleAuditor
	default:
		return models.RoleUser
	}
}

func AuthMiddleware(authn Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		id, err := authn.Authenticate(c.Request.Context(), tokenString)
		if errors.Is(err, userservice.ErrUpstream) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service unavailable"})
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		userID, err := strconv.ParseInt(id.UserID, 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		role := kmsRole(id)
		c.Set("user_id", userID)
		c.Set("username", id.Username)
		c.Set("role", role)
		c.Set("permissions", rolePermissions[role])

		c.Next()
	}
}

func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		for _, requiredRole := range roles {
			if role == requiredRole || role == models.RoleAdmin {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permissions not found"})
			c.Abort()
			return
		}

		perms := permissions.([]string)
		for _, perm := range perms {
			if perm == permission || perm == "*" {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		c.Abort()
	}
}
