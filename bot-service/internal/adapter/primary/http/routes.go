package http

import (
	"github.com/JIeeiroSst/bot-service/config"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler, fb *FacebookHandler, content *ContentHandler, cfg *config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", h.GetHealth)

	posts := engine.Group("/bot-content/v1/posts")
	{
		posts.POST("", content.CreatePost)
		posts.GET("", content.ListPosts)
		posts.GET("/:id", content.GetPost)
		posts.PATCH("/:id", content.UpdatePost)
		posts.POST("/:id/submit", content.SubmitForReview)
		posts.POST("/:id/approve", content.ApprovePost)
		posts.POST("/:id/reject", content.RejectPost)
		posts.POST("/:id/publish", content.PublishNow)
		posts.POST("/:id/cancel", content.CancelPost)
	}

	if cfg.Facebook.Enabled {
		engine.GET("/webhook/facebook", fb.VerifyWebhook)
		engine.POST("/webhook/facebook", fb.ReceiveWebhook)
	}

	return engine
}
