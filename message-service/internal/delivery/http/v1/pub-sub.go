package v1

import "github.com/gin-gonic/gin"

func (h *Handler) initMessageRoutes(api *gin.RouterGroup) {
	group := api.Group("/queue")

	group.POST("/pub", h.middleware.AuthorizeControl(), h.Producer)
	group.POST("/sub", h.middleware.AuthorizeControl(), h.Consume)
}

func (h *Handler) Producer(ctx *gin.Context) {

}

func (h *Handler) Consume(ctx *gin.Context) {

}
