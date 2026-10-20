package middlware

import (
	"strings"

	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/gin-gonic/gin"
)

func ClientInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		device := strings.TrimSpace(c.GetHeader("X-Device-Id"))
		if len(device) > 128 {
			device = device[:128]
		}
		ctx := auth.WithClient(c.Request.Context(), auth.Client{IP: c.ClientIP(), UserAgent: c.Request.UserAgent(), DeviceID: device})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
