package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	basket         port.BasketUsecase
	basketLine     port.BasketLineUsecase
	basketLineAttr port.BasketLineAttributeUsecase
	order          port.OrderUsecase
	user           port.UserUsecase
	checkout       port.CheckoutUsecase
}

func NewHandler(
	basket port.BasketUsecase,
	basketLine port.BasketLineUsecase,
	basketLineAttr port.BasketLineAttributeUsecase,
	order port.OrderUsecase,
	user port.UserUsecase,
	checkout port.CheckoutUsecase,
) *Handler {
	return &Handler{
		basket:         basket,
		basketLine:     basketLine,
		basketLineAttr: basketLineAttr,
		order:          order,
		user:           user,
		checkout:       checkout,
	}
}

// writeError translates domain errors into the matching HTTP status code.
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrOutOfStock), errors.Is(err, common.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrPaymentFailed):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrUpstream):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
