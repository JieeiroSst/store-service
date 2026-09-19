package http

import (
	"net/http"

	"github.com/JIeeiroSst/threads-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FollowUser(c *gin.Context) {
	if err := h.follows.Follow(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UnfollowUser(c *gin.Context) {
	if err := h.follows.Unfollow(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListFollowers(c *gin.Context) {
	cursor, limit := listParams(c)
	authors, next, err := h.follows.ListFollowers(c.Request.Context(), c.Param("id"), cursor, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewPage(authors, next))
}

func (h *Handler) ListFollowing(c *gin.Context) {
	cursor, limit := listParams(c)
	authors, next, err := h.follows.ListFollowing(c.Request.Context(), c.Param("id"), cursor, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewPage(authors, next))
}
