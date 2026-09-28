package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	products port.ProductUsecase
	orders   port.OrderUsecase
	webhooks port.WebhookUsecase
}

func NewHandler(products port.ProductUsecase, orders port.OrderUsecase, webhooks port.WebhookUsecase) *Handler {
	return &Handler{products: products, orders: orders, webhooks: webhooks}
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInvalidWebhook):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrShopifyUserError):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrShopifyRequestFailed):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func parseList(c *gin.Context) port.ListInput {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	return port.ListInput{Limit: limit, Offset: offset}
}
