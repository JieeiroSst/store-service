package http

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.MaxMultipartMemory = 8 << 20

	r.GET("/health", h.Health)
	r.POST("/upload", h.Upload)
	r.GET("/images/:id", h.GetImage)
	r.GET("/images/:id/info", h.GetInfo)
	return r
}
