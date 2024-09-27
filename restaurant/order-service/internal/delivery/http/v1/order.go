package v1

import "github.com/gin-gonic/gin"

func (h *Handler) initOrderRoutes(api *gin.RouterGroup) {
	group := api.Group("/order")

	group.GET("/:id", h.FindByID)
	group.GET("", h.FindAll)
}

func (h *Handler) FindByID(ctx *gin.Context) {

}

func (h *Handler) FindAll(ctx *gin.Context) {

}
