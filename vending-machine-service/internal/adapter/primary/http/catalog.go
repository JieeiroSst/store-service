package http

import (
	"net/http"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateCategory(c *gin.Context) {
	var req categoryRequest
	if !bind(c, &req) {
		return
	}
	cat, err := h.catalog.CreateCategory(c.Request.Context(), &domain.Category{
		Name:         req.Name,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toCategory(cat))
}

func (h *Handler) ListCategories(c *gin.Context) {
	page, err := pageRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	cats, err := h.catalog.ListCategories(c.Request.Context(), page)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPage(cats, page, toCategory))
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req productRequest
	if !bind(c, &req) {
		return
	}
	active := req.IsActive == nil || *req.IsActive
	p, err := h.catalog.CreateProduct(c.Request.Context(), &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
		Barcode:     req.Barcode,
		IsActive:    active,
		Attributes:  req.Attributes,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toProduct(p))
}

func (h *Handler) ListProducts(c *gin.Context) {
	page, err := pageRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	products, err := h.catalog.ListProducts(c.Request.Context(), c.Query("category_id"), page)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPage(products, page, toProduct))
}

func (h *Handler) GetProduct(c *gin.Context) {
	p, err := h.catalog.GetProduct(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProduct(p))
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	var req productPatchRequest
	if !bind(c, &req) {
		return
	}
	p, err := h.catalog.UpdateProduct(c.Request.Context(), c.Param("id"), domain.ProductPatch{
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		ImageURL:    req.ImageURL,
		Barcode:     req.Barcode,
		IsActive:    req.IsActive,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProduct(p))
}
