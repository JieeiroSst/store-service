package http

import (
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type pageResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	IsLastPage bool   `json:"is_last_page"`
	Limit      int    `json:"limit"`
}

func pageRequest(c *gin.Context) (domain.PageRequest, error) {
	page := domain.PageRequest{Cursor: c.Query("cursor")}
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > domain.MaxPageSize {
			return page, fmt.Errorf("%w: limit must be between 1 and %d", domain.ErrInvalidInput, domain.MaxPageSize)
		}
		page.Limit = n
	}
	return page, nil
}

func toPage[T, R any](p domain.Page[T], req domain.PageRequest, f func(*T) R) pageResponse[R] {
	return pageResponse[R]{
		Items:      mapSlice(p.Items, f),
		NextCursor: p.NextCursor,
		IsLastPage: p.IsLastPage,
		Limit:      req.Size(),
	}
}
