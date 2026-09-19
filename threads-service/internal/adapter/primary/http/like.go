package http

import (
	"net/http"

	"github.com/JIeeiroSst/threads-service/internal/adapter/primary/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) LikePost(c *gin.Context) {
	if err := h.likes.LikePost(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UnlikePost(c *gin.Context) {
	if err := h.likes.UnlikePost(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) LikeComment(c *gin.Context) {
	if err := h.likes.LikeComment(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UnlikeComment(c *gin.Context) {
	if err := h.likes.UnlikeComment(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
