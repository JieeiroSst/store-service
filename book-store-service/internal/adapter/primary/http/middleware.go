package http

import (
	"github.com/JIeeiroSst/bookStore-service/utils"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type,"+
				"Content-Length, "+
				"Accept-Encoding, "+
				"X-CSRF-Token, "+
				"Authorization, "+
				"accept, "+
				"origin, "+
				"Cache-Control, "+
				"X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// AuthorizeMiddleware validates the Authorization header against the
// configured secret. Not wired into any route group by default — kept
// available for handlers that need to opt into it, mirroring the previous
// (unused) middleware.Middleware from before the hexagonal migration.
func AuthorizeMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.Request.Header.Get("Authorization")
		if ok := utils.DecodeBase(authorization, secret); !ok {
			c.AbortWithStatusJSON(403, gin.H{"message": "Unauthorized"})
			return
		}
		c.Next()
	}
}
