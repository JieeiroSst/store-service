package middleware

import (
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
