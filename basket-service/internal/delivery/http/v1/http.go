package v1

import (
	"github.com/JIeeiroSst/basket-service/internal/usecase"
	"github.com/JIeeiroSst/basket-service/middleware"
	"github.com/gin-gonic/gin"
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

func (h *Handler) Init(api *gin.RouterGroup) {
	_ = api.Group("/v1")
	{
		// h.initRoutes(v1)
	}
}
