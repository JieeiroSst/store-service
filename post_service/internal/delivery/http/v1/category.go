package v1

import (
	"github.com/JIeeiroSst/post-service/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initCategoryRouter(api *gin.RouterGroup) {
	categoryGroup := api.Group("/category")
	categoryGroup.GET("/", h.Categories)
	categoryGroup.GET("/:id", h.CategoryById)
	categoryGroup.POST("/", h.CreateCategory)
	categoryGroup.PUT("/:id", h.UpdateCategory)
	categoryGroup.DELETE("/:id", h.DeleteCategory)
}

func (h *Handler) CreateCategory(ctx *gin.Context) {
	var category model.Category
	if err := ctx.ShouldBind(&category); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := h.usecase.Categories.Create(category); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "create gory success",
	})
}

func (h *Handler) UpdateCategory(ctx *gin.Context) {
	id := ctx.Param("id")

	var category model.Category
	if err := ctx.ShouldBind(&category); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := h.usecase.Categories.Update(id, category); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "update category success",
	})
}

func (h *Handler) DeleteCategory(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := h.usecase.Categories.Delete(id); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "delete category success",
	})
}

func (h *Handler) Categories(ctx *gin.Context) {
	categories, err := h.usecase.Categories.Categories()
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, categories)
}

func (h *Handler) CategoryById(ctx *gin.Context) {
	id := ctx.Param("id")

	category, err := h.usecase.Categories.CategoryById(id)
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
	}
	ctx.JSON(200, category)
}
