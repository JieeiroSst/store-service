package http

import (
	"net/http"

	"github.com/JIeeiroSst/threads-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Bookmark(c *gin.Context) {
	if err := h.bookmarks.Bookmark(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Unbookmark(c *gin.Context) {
	if err := h.bookmarks.Unbookmark(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListBookmarks returns the caller's own saved posts - never anyone
// else's, since bookmarks are private (unlike likes/reposts/follows).
func (h *Handler) ListBookmarks(c *gin.Context) {
	cursor, limit := listParams(c)
	posts, next, err := h.bookmarks.ListBookmarks(c.Request.Context(), middleware.UserID(c), cursor, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewPage(posts, next))
}
