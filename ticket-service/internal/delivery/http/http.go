package http

import (
	v1 "github.com/JIeeiroSst/ticket-service/internal/delivery/http/v1"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Init(router *gin.Engine) {
	h.corsMiddleware(router)
	h.initApi(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler()
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
