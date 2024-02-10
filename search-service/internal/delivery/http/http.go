package http

import (
	v1 "github.com/JIeeiroSst/search-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/search-service/internal/service"
	"github.com/JIeeiroSst/search-service/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	middleware middleware.Middleware
	service    service.Service
}

func NewHandler(middleware middleware.Middleware) *Handler {
	return &Handler{
		middleware: middleware,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.corsMiddleware(router)
	h.initApi(router)
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.service, h.middleware)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
