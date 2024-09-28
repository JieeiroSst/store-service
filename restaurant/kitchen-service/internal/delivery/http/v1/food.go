package v1

import (
	"github.com/JIeeiroSst/kitchen-service/internal/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initFoodRoutes(api *gin.RouterGroup) {
	group := api.Group("/food")

	group.GET("", h.FindFood)
	group.POST("", h.CreateFood)
}

func (h *Handler) FindFood(ctx *gin.Context) {
	orders, err := h.usecase.Categories.Find(ctx)
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

func (h *Handler) CreateFood(ctx *gin.Context) {
	var food dto.Food
	if err := ctx.ShouldBind(&food); err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}

	if err := h.usecase.Foods.Create(ctx, food); err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
	})
}
