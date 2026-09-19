package http

import (
	"net/http"

	"github.com/JIeeiroSst/post-service/model"
	"github.com/gin-gonic/gin"
)

type categoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.categories.CreateCategory(c.Request.Context(), model.CreateCategoryInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.categories.UpdateCategory(c.Request.Context(), c.Param("id"), model.UpdateCategoryInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category updated"})
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	if err := h.categories.DeleteCategory(c.Request.Context(), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListCategories is cursor-paginated - see listParams.
func (h *Handler) ListCategories(c *gin.Context) {
	cursor, limit := listParams(c)
	categories, next, err := h.categories.ListCategories(c.Request.Context(), cursor, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewPage(categories, next))
}

func (h *Handler) GetCategory(c *gin.Context) {
	category, err := h.categories.GetCategory(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, category)
}
