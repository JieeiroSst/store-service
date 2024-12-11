package v1

import (
	"github.com/JIeeiroSst/ticket-service/internal/usecase"
	"github.com/JIeeiroSst/ticket-service/middleware"
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
		h.initInvoices(v1)
		h.initTickets(v1)
	}
}
