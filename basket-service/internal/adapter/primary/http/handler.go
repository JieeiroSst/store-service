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
}

func NewHandler(
	basket port.BasketUsecase,
	basketLine port.BasketLineUsecase,
	basketLineAttr port.BasketLineAttributeUsecase,
	order port.OrderUsecase,
	user port.UserUsecase,
) *Handler {
	return &Handler{
		basket:         basket,
		basketLine:     basketLine,
		basketLineAttr: basketLineAttr,
		order:          order,
		user:           user,
	}
}

// writeError translates domain errors into the matching HTTP status code.
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
