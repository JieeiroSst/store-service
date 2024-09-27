package http

import (
	v1 "github.com/JIeeiroSst/order-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/order-service/internal/usecase"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.initApi(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h.corsMiddleware(router)
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(&h.usecase)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
