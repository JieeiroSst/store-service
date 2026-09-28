package http

import (
	"net/http"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createProductRequest struct {
	Title           string   `json:"title" binding:"required"`
	DescriptionHTML string   `json:"description_html"`
	Vendor          string   `json:"vendor"`
	ProductType     string   `json:"product_type"`
	Tags            []string `json:"tags"`
	Status          string   `json:"status"`
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.products.CreateProduct(c.Request.Context(), port.CreateProductInput{
		Title:           req.Title,
		DescriptionHTML: req.DescriptionHTML,
		Vendor:          req.Vendor,
		ProductType:     req.ProductType,
		Tags:            req.Tags,
		Status:          model.ProductStatus(req.Status),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, product)
}

func (h *Handler) GetProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	product, err := h.products.GetProduct(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *Handler) ListProducts(c *gin.Context) {
	list, err := h.products.ListProducts(c.Request.Context(), parseList(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list.Items, "total": list.Total})
}

func (h *Handler) SyncProducts(c *gin.Context) {
	result, err := h.products.SyncProducts(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
