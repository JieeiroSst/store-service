package v1

import (
	"strconv"

	"github.com/JieeiroSst/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initOrderRoutes(api *gin.RouterGroup) {
	group := api.Group("/order")

	group.GET("/:id", h.FindByID)
	group.GET("", h.FindAll)
}

func (h *Handler) FindByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}

	order, err := h.usecase.FindByID(ctx, id)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"data": order,
	})
}

func (h *Handler) FindAll(ctx *gin.Context) {
	var pagination logger.Pagination
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}
	orders, err := h.usecase.FindAll(ctx, pagination)
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
