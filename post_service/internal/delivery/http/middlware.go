package http

import (
	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) corsMiddleware(router *gin.Engine) {
	router.Use(cors.Default())
}