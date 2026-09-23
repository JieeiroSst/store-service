package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1", h.authenticate)
	{
		api.GET("/categories", h.Categories)
		api.GET("/leaderboard", h.Leaderboard)

		events := api.Group("/events")
		{
			events.GET("", h.ListEvents)
			events.POST("", h.requireAdmin, h.CreateEvent)
			events.GET("/:id", h.GetEvent)
			events.GET("/:id/related", h.RelatedEvents)
			events.GET("/:id/comments", h.ListComments)
			events.POST("/:id/comments", h.AddComment)
			events.POST("/:id/convert", h.ConvertPositions)
			events.POST("/:id/propose", h.requireAdmin, h.ProposeEvent)
			events.POST("/:id/resolve", h.requireAdmin, h.ResolveEvent)
			events.POST("/:id/watch", h.Watch)
			events.DELETE("/:id/watch", h.Unwatch)
		}

		comments := api.Group("/comments/:commentId")
		{
			comments.DELETE("", h.DeleteComment)
			comments.POST("/like", h.LikeComment)
			comments.DELETE("/like", h.UnlikeComment)
		}

		markets := api.Group("/markets")
		{
			markets.GET("", h.SearchMarkets)
			markets.GET("/:id", h.GetMarket) // numeric id or slug
			markets.GET("/:id/book", h.GetBook)
			markets.GET("/:id/quote", h.GetQuote)
			markets.GET("/:id/trades", h.ListMarketTrades)
			markets.GET("/:id/prices-history", h.GetPriceHistory)
			markets.GET("/:id/holders", h.TopHolders)
			markets.GET("/:id/stream", h.Stream)
			markets.GET("/:id/rewards", h.GetMarketRewards)
			markets.POST("/:id/rewards", h.requireAdmin, h.SetMarketRewards)

			markets.POST("/:id/orders", h.PlaceOrder)
			markets.DELETE("/:id/orders", h.CancelAllOrders)

			markets.POST("/:id/propose", h.requireAdmin, h.ProposeResolution)
			markets.POST("/:id/dispute", h.DisputeResolution)
			markets.POST("/:id/finalize", h.FinalizeResolution)
			markets.POST("/:id/resolve", h.requireAdmin, h.ResolveMarket)
		}

		admin := api.Group("/admin", h.requireAdmin)
		{
			admin.GET("/exchange", h.GetExchange)
			admin.POST("/exchange/fund", h.FundExchange)
		}

		orders := api.Group("/orders")
		{
			orders.GET("/:orderId", h.GetOrder)
			orders.DELETE("/:orderId", h.CancelOrder)
		}

		users := api.Group("/users/:userId")
		{
			users.GET("/profile", h.GetProfile)
			users.PUT("/profile", h.requireOwner, h.UpdateProfile)

			private := users.Group("", h.requireOwner)
			private.GET("/balance", h.GetBalance)
			private.GET("/wallet", h.GetWallet)
			private.POST("/deposit", h.Deposit)
			private.POST("/withdraw", h.Withdraw)
			private.GET("/portfolio", h.GetPortfolio)
			private.GET("/positions/closed", h.ClosedPositions)
			private.GET("/activity", h.GetActivity)
			private.GET("/orders", h.ListUserOrders)
			private.GET("/watchlist", h.Watchlist)
			private.GET("/rewards", h.GetUserRewards)
			private.GET("/referral", h.GetReferral)
			private.POST("/referral/redeem", h.RedeemReferral)
		}
	}

	return engine
}
