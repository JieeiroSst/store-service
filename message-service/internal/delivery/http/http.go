package http

import (
	v1 "github.com/JIeeiroSst/message-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/message-service/internal/usecase"
	"github.com/JIeeiroSst/message-service/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	middleware middleware.Middleware
	usecase    usecase.Usecase
}

func NewHandler(middleware middleware.Middleware, usecase usecase.Usecase) *Handler {
	return &Handler{
		middleware: middleware,
		usecase:    usecase,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.corsMiddleware(router)
	h.initApi(router)
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.middleware, h.usecase)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
