package http

import (
	"net/http"

	"github.com/JIeeiroSst/post-service/config"
	"github.com/JIeeiroSst/post-service/internal/adapter/primary/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler, cfg *config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), cors.Default())

	engine.GET("/health", getHealth)

	auth := middleware.RequireAuth(cfg.Auth.JWTSecret)

	api := engine.Group("/api/v1")
	{
		posts := api.Group("/post")
		{
			// Public reads - browsing posts shouldn't require an account.
			posts.GET("", h.ListPosts)
			posts.GET("/:id", h.GetPost)

			// Authenticated writes - the original routes had ListPosts
			// bound to "/:id" and GetPost bound to "/" (swapped), and
			// CreatePost read a "category-id" URL param the route never
			// declared, so it was always empty.
			posts.POST("", auth, h.CreatePost)
			posts.PUT("/:id", auth, h.UpdatePost)
		}

		categories := api.Group("/category")
		{
			categories.GET("", h.ListCategories)
			categories.GET("/:id", h.GetCategory)
			categories.POST("", auth, h.CreateCategory)
			categories.PUT("/:id", auth, h.UpdateCategory)
			categories.DELETE("/:id", auth, h.DeleteCategory)
		}
	}

	return engine
}
