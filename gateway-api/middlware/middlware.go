package middlware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewMiddlware() gin.HandlerFunc {
	config := cors.Default()

	return config
}
