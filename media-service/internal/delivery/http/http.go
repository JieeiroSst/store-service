package http

import (
	v1 "github.com/JIeeiroSst/media-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/media-service/internal/usecase"
	"github.com/JIeeiroSst/media-service/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	usecase    *usecase.Usecase
	middleware middleware.Middleware
}

func NewHandler(usecase *usecase.Usecase, middleware middleware.Middleware) *Handler {
	return &Handler{
		usecase:    usecase,
		middleware: middleware,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.corsMiddleware(router)
	h.initApi(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.usecase, h.middleware)
	apiInternal := router.Group("/api/internal")
	{
		handlerV1.Init(apiInternal)
	}
}
