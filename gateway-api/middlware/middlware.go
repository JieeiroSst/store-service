package middlware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewMiddlware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")

	return cors.New(config)
}
