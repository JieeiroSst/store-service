package http

import (
	v1 "github.com/JIeeiroSst/accounting-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/accounting-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	nats    *nats.Conn
	usecase *usecase.Usecase
}

func NewHandler(nats *nats.Conn, usecase *usecase.Usecase) *Handler {
	return &Handler{
		nats:    nats,
		usecase: usecase,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.initApi(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h.corsMiddleware(router)
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.nats, h.usecase)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
