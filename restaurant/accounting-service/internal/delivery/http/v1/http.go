package v1

import (
	"github.com/JIeeiroSst/accounting-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
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

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initAuthCartRoutes(v1)
	}
}
