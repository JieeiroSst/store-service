package v1

import (
	"github.com/JIeeiroSst/message-service/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	middleware middleware.Middleware
}

func NewHandler(middleware middleware.Middleware) *Handler {
	return &Handler{
		middleware: middleware,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initMessageRoutes(v1)
	}
}
