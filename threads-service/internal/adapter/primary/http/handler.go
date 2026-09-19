package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	posts     port.PostUsecase
	comments  port.CommentUsecase
	likes     port.LikeUsecase
	follows   port.FollowUsecase
	bookmarks port.BookmarkUsecase
	tags      port.TagUsecase
}

func NewHandler(posts port.PostUsecase, comments port.CommentUsecase, likes port.LikeUsecase, follows port.FollowUsecase, bookmarks port.BookmarkUsecase, tags port.TagUsecase) *Handler {
	return &Handler{posts: posts, comments: comments, likes: likes, follows: follows, bookmarks: bookmarks, tags: tags}
}

// listParams reads the two query params every cursor-paginated list
// endpoint accepts: an opaque cursor (empty means "first page") and a
// limit (0/invalid falls through to model.ClampLimit's default at the
// repository layer - not parsed/defaulted here, so there's exactly one
// place that decision is made).
func listParams(c *gin.Context) (cursor string, limit int) {
	cursor = c.Query("cursor")
	limit, _ = strconv.Atoi(c.Query("limit"))
	return cursor, limit
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrAlreadyLiked), errors.Is(err, port.ErrAlreadyFollowing),
		errors.Is(err, port.ErrAlreadyReposted), errors.Is(err, port.ErrAlreadyBookmarked):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrCannotFollowSelf):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
