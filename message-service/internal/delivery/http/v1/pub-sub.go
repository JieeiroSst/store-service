package v1

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initMessageRoutes(api *gin.RouterGroup) {
	group := api.Group("/queue")

	group.POST("/pub", h.middleware.AuthorizeControl(), h.Producer)
	group.POST("/sub", h.middleware.AuthorizeControl(), h.Consume)
}

func (h *Handler) Producer(ctx *gin.Context) {
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Error reading request body")
		return
	}

	topic := ctx.Query("topic")

	if err := h.usecase.PubSub.Producer(ctx, topic, data); err != nil {
		ctx.JSON(500, "Internal Server Error")
		return
	}

	ctx.JSON(200, "successfully")
}

func (h *Handler) Consume(ctx *gin.Context) {
	topic := ctx.Query("topic")

	data, err := h.usecase.PubSub.Consume(ctx, topic)
	if err != nil {
		ctx.JSON(500, "Internal Server Error")
		return
	}

	ctx.JSON(200, string(data))
}
