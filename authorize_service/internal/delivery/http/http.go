package http

import (
	_ "github.com/JieeiroSst/authorize-service/docs"
	v1 "github.com/JieeiroSst/authorize-service/internal/delivery/http/v1"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/middleware"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	usecase    usecase.Usecase
	middleware middleware.Middleware
}

func NewHandler(usecase usecase.Usecase, middleware middleware.Middleware, adapter persist.Adapter) *Handler {
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
	handlerV1 := v1.NewHandler(&h.usecase, h.middleware)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
