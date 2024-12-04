package v1

import (
	"github.com/JIeeiroSst/calculate-service/internal/usecase"
	"github.com/JIeeiroSst/calculate-service/middleware"
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
		h.initCampaignConfig(v1)
		h.initUser(v1)
	}
}
