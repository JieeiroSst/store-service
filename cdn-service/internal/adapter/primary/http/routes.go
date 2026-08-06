package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", h.GetHealth)
	engine.StaticFS("/media", http.Dir(h.baseUpload.BaseDirUpload))
	engine.GET("/v1/files/:file_id/content", h.GetFile)

	return engine
}
