package middleware

import (
	"errors"
	"net/http"
	"strings"

	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	"github.com/gin-gonic/gin"
)

const claimsContextKey = "auth_claims"

func Auth(authSvc authapp.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "missing_token", "message": "authorization header required"}})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")

		claims, err := authSvc.VerifyToken(c.Request.Context(), token)
		if err != nil && !errors.Is(err, authapp.ErrUnauthenticated) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "auth_unavailable", "message": "authentication service unavailable"}})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "invalid_token", "message": "invalid or expired token"}})
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func ClaimsFromGin(c *gin.Context) (*authapp.Claims, bool) {
	v, ok := c.Get(claimsContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*authapp.Claims)
	return claims, ok
}

func IsAdmin(c *gin.Context) bool {
	claims, ok := ClaimsFromGin(c)
	return ok && (claims.Role == "admin" || claims.Role == "super_admin")
}

func CallerID(c *gin.Context) string {
	if claims, ok := ClaimsFromGin(c); ok {
		return claims.UserID
	}
	return ""
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "admin role required"}})
			return
		}
		c.Next()
	}
}
