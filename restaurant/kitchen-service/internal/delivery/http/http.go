package http

import (
	v1 "github.com/JIeeiroSst/kitchen-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/kitchen-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	usecase *usecase.Usecase
	nats    *nats.Conn
}

func NewHandler(usecase *usecase.Usecase, nats *nats.Conn) *Handler {
	return &Handler{
		usecase: usecase,
		nats:    nats,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.initApi(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h.corsMiddleware(router)
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.usecase, h.nats)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
