package v1

import (
	"github.com/JIeeiroSst/kitchen-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
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

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initCategoryRoutes(v1)
		h.initFoodRoutes(v1)
		h.initKitchenRoutes(v1)
	}
}
