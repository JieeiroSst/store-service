package v1

import (
	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initKitchenRoutes(api *gin.RouterGroup) {
	group := api.Group("/kitchen")

	group.GET("", h.FindKitchen)
}

func (h *Handler) FindKitchen(ctx *gin.Context) {
	var pagination logger.Pagination
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}
	orders, err := h.usecase.Kitchens.Find(ctx, pagination)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"data": orders,
	})
}
