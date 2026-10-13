package middlware

import (
	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/gin-gonic/gin"
)

func ClientInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := auth.WithClient(c.Request.Context(), auth.Client{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
