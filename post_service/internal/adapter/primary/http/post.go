package http

import (
	"net/http"

	"github.com/JIeeiroSst/post-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
	"github.com/gin-gonic/gin"
)

type createPostRequest struct {
	Name        string `form:"name" binding:"required"`
	Content     string `form:"content"`
	Description string `form:"description"`
	CategoryID  string `form:"category_id"`
}

// CreatePost takes the author from the verified JWT, never the request
// body - otherwise any caller could post "as" someone else. The uploaded
// file is optional: a text-only post is valid.
func (h *Handler) CreatePost(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var file *port.UploadFileInput
	if fileHeader, err := c.FormFile("file"); err == nil && fileHeader != nil {
		opened, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer opened.Close()

		contentType := fileHeader.Header.Get("Content-Type")
		file = &port.UploadFileInput{
			FileName:    fileHeader.Filename,
			ContentType: contentType,
			Size:        fileHeader.Size,
			Reader:      opened,
		}
	}

	post, err := h.posts.CreatePost(c.Request.Context(), model.CreatePostInput{
		AuthorID:    middleware.UserID(c),
		Name:        req.Name,
		Content:     req.Content,
		Description: req.Description,
		CategoryID:  req.CategoryID,
	}, file)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, post)
}

// ListPosts is cursor-paginated - see listParams.
func (h *Handler) ListPosts(c *gin.Context) {
	cursor, limit := listParams(c)
	posts, next, err := h.posts.ListPosts(c.Request.Context(), cursor, limit)
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

type updatePostRequest struct {
	Name        string `json:"name" binding:"required"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

// UpdatePost is author-only, enforced inside the usecase (see
// application.postService.UpdatePost) - the original handler let any
// caller edit any post.
func (h *Handler) UpdatePost(c *gin.Context) {
	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.posts.UpdatePost(c.Request.Context(), c.Param("id"), middleware.UserID(c), model.UpdatePostInput{
		Name:        req.Name,
		Content:     req.Content,
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post updated"})
}
