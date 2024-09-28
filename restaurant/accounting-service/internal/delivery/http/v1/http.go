package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type Handler struct {
	nats *nats.Conn
}

func NewHandler(nats *nats.Conn) *Handler {
	return &Handler{
		nats: nats,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initAuthCartRoutes(v1)
	}
}
