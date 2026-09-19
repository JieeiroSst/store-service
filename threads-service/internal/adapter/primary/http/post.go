package http

import (
	"net/http"

	"github.com/JIeeiroSst/threads-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

type createPostRequest struct {
	Content   string   `json:"content" binding:"required"`
	MediaURLs []string `json:"media_urls"`
	Tags      []string `json:"tags"`
}

func (h *Handler) CreatePost(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.posts.CreatePost(c.Request.Context(), model.CreatePostInput{
		UserID:    middleware.UserID(c),
		Content:   req.Content,
		MediaURLs: req.MediaURLs,
		Tags:      req.Tags,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, post)
}

// ListFeed returns every post, or just one author's posts (a profile grid)
// when ?author_id= is set. Cursor-paginated - see listParams.
func (h *Handler) ListFeed(c *gin.Context) {
	cursor, limit := listParams(c)
	posts, next, err := h.posts.ListFeed(c.Request.Context(), c.Query("author_id"), cursor, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewPage(posts, next))
}

func (h *Handler) GetPost(c *gin.Context) {
	post, err := h.posts.GetPost(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *Handler) DeletePost(c *gin.Context) {
	if err := h.posts.DeletePost(c.Request.Context(), c.Param("id"), middleware.UserID(c)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListHomeFeed is the caller's "following" timeline, as opposed to
// ListFeed's global/profile views.
func (h *Handler) ListHomeFeed(c *gin.Context) {
	cursor, limit := listParams(c)
	posts, next, err := h.posts.ListHomeFeed(c.Request.Context(), middleware.UserID(c), cursor, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewPage(posts, next))
}

func (h *Handler) Repost(c *gin.Context) {
	repost, err := h.posts.Repost(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, repost)
}

func (h *Handler) Unrepost(c *gin.Context) {
	if err := h.posts.Unrepost(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
