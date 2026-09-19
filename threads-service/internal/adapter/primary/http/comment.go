package http

import (
	"net/http"

	"github.com/JIeeiroSst/threads-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

type createCommentRequest struct {
	Content         string   `json:"content" binding:"required"`
	MediaURLs       []string `json:"media_urls"`
	ParentCommentID *string  `json:"parent_comment_id"`
}

func (h *Handler) CreateComment(c *gin.Context) {
	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.comments.CreateComment(c.Request.Context(), model.CreateCommentInput{
		PostID:          c.Param("id"),
		UserID:          middleware.UserID(c),
		ParentCommentID: req.ParentCommentID,
		Content:         req.Content,
		MediaURLs:       req.MediaURLs,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// ListComments returns a flat, oldest-first, cursor-paginated list;
// parent_comment_id lets the client build the reply tree instead of the
// server nesting JSON.
func (h *Handler) ListComments(c *gin.Context) {
	cursor, limit := listParams(c)
	comments, next, err := h.comments.ListComments(c.Request.Context(), c.Param("id"), cursor, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewPage(comments, next))
}

func (h *Handler) DeleteComment(c *gin.Context) {
	if err := h.comments.DeleteComment(c.Request.Context(), c.Param("id"), middleware.UserID(c)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
