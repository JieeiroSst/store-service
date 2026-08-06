package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/bookStore-service/common"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	book      port.BookUsecase
	author    port.AuthorUsecase
	publisher port.PublisherUsecase
	category  port.CategoryUsecase
}

func NewHandler(
	book port.BookUsecase,
	author port.AuthorUsecase,
	publisher port.PublisherUsecase,
	category port.CategoryUsecase,
) *Handler {
	return &Handler{
		book:      book,
		author:    author,
		publisher: publisher,
		category:  category,
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
