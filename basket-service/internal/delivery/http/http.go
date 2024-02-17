package http

import (
	v1 "github.com/JIeeiroSst/basket-service/internal/delivery/http/v1"
	"github.com/JIeeiroSst/basket-service/internal/usecase"
	"github.com/JIeeiroSst/basket-service/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase    usecase.Usecase
	middleware middleware.Middleware
}

func NewHandler(usecase usecase.Usecase, middleware middleware.Middleware) *Handler {
	return &Handler{
		usecase:    usecase,
		middleware: middleware,
	}
}

func (h *Handler) Init(router *gin.Engine) {
	h.corsMiddleware(router)
	h.initApi(router)
}

func (h *Handler) initApi(router *gin.Engine) {
	handlerV1 := v1.NewHandler(&h.usecase, h.middleware)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
