package v1

import (
	"github.com/JIeeiroSst/payer-service/internal/usecase"
	"github.com/JIeeiroSst/payer-service/middleware"
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
	v1 := api.Group("/v1")
	{
		h.initTransaction(v1)
	}
}
