package v1

import (
	"github.com/JIeeiroSst/kitchen-service/internal/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initCategoryRoutes(api *gin.RouterGroup) {
	group := api.Group("/category")

	group.GET("", h.FindCategory)
	group.POST("", h.CreateCategory)
}

func (h *Handler) FindCategory(ctx *gin.Context) {
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

func (h *Handler) CreateCategory(ctx *gin.Context) {
	var category dto.Category
	if err := ctx.ShouldBind(&category); err != nil {
		ctx.JSON(400, gin.H{
			"error": err,
		})
		return
	}

	if err := h.usecase.Categories.Create(ctx, category); err != nil {
		ctx.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
	})
}
