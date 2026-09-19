package http

import (
	"net/http"

	"github.com/JIeeiroSst/threads-service/config"
	"github.com/JIeeiroSst/threads-service/internal/adapter/primary/http/middleware"
	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler, cfg *config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	auth := middleware.RequireAuth(cfg.Auth.JWTSecret)

	api := engine.Group("/api/v1")
	{
		posts := api.Group("/posts")
		{
			// Public reads - browsing a feed shouldn't require an account.
			posts.GET("", h.ListFeed)
			posts.GET("/:id", h.GetPost)
			posts.GET("/:id/comments", h.ListComments)

			// Authenticated writes - ownership of the specific post is
			// enforced inside the usecase, not just here.
			posts.POST("", auth, h.CreatePost)
			posts.DELETE("/:id", auth, h.DeletePost)
			posts.POST("/:id/likes", auth, h.LikePost)
			posts.DELETE("/:id/likes", auth, h.UnlikePost)
			posts.POST("/:id/comments", auth, h.CreateComment)
			posts.POST("/:id/repost", auth, h.Repost)
			posts.DELETE("/:id/repost", auth, h.Unrepost)
			posts.POST("/:id/bookmark", auth, h.Bookmark)
			posts.DELETE("/:id/bookmark", auth, h.Unbookmark)
		}

		comments := api.Group("/comments")
		{
			comments.DELETE("/:id", auth, h.DeleteComment)
			comments.POST("/:id/likes", auth, h.LikeComment)
			comments.DELETE("/:id/likes", auth, h.UnlikeComment)
		}

		users := api.Group("/users")
		{
			users.GET("/:id/followers", h.ListFollowers)
			users.GET("/:id/following", h.ListFollowing)
			users.POST("/:id/follow", auth, h.FollowUser)
			users.DELETE("/:id/follow", auth, h.UnfollowUser)
		}

		// The caller's "following" timeline and private saved posts - both
		// scoped to the JWT's subject, never another user's, so there's no
		// :id param to authorize against.
		feed := api.Group("/feed", auth)
		{
			feed.GET("", h.ListHomeFeed)
		}
		bookmarks := api.Group("/bookmarks", auth)
		{
			bookmarks.GET("", h.ListBookmarks)
		}

		tags := api.Group("/tags")
		{
			tags.GET("/trending", h.ListTrendingTags)
		}
	}

	return engine
}
