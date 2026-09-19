package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	posts      port.PostUsecase
	categories port.CategoryUsecase
}

func NewHandler(posts port.PostUsecase, categories port.CategoryUsecase) *Handler {
	return &Handler{posts: posts, categories: categories}
}

// writeError maps a usecase error to the right HTTP status - callers of
// this API need 403 vs 404 vs 500 to behave sensibly, not a blanket 500
// for everything (the original handlers returned 400/500 for every
// failure, not-found included).
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// listParams reads the two query params every cursor-paginated list
// endpoint accepts: an opaque cursor (empty means "first page") and a
// limit (0/invalid falls through to model.ClampLimit's default at the
// repository layer).
func listParams(c *gin.Context) (cursor string, limit int) {
	cursor = c.Query("cursor")
	limit, _ = strconv.Atoi(c.Query("limit"))
	return cursor, limit
}
