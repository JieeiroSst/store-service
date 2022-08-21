package v1

import (
	"github.com/JIeeiroSst/post-service/model"
	"github.com/JIeeiroSst/post-service/pkg/minio"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initPostRoutes(api *gin.RouterGroup) {
	postGroup := api.Group("/post")
	postGroup.POST("/", h.CreatePost)
	postGroup.GET("/:id", h.Posts)
	postGroup.GET("/", h.PostById)
	postGroup.PUT("/:id", h.UpdatePost)
}

func (h *Handler) CreatePost(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	upload := minio.UploadObjectArgs{
		FileHeader: file,
	}
	var new model.New
	if err := ctx.ShouldBind(&new); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	cateId := ctx.Param("category-id")

	if err := h.usecase.News.Create(cateId, new, upload); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "create post success",
	})
}

func (h *Handler) Posts(ctx *gin.Context) {
	posts, err := h.usecase.News.News()
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, posts)
}

func (h *Handler) PostById(ctx *gin.Context) {
	id := ctx.Param("id")
	post, err := h.usecase.News.NewById(id)
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, post)
}

func (h *Handler) UpdatePost(ctx *gin.Context) {
	id := ctx.Param("id")
	var new model.New
	if err := ctx.ShouldBind(&new); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := h.usecase.News.Update(id, new); err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "update post success",
	})
}
