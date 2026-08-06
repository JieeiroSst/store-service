package http

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), CORSMiddleware())

	engine.GET("/health", h.GetHealth)

	api := engine.Group("/api/v1")
	RegisterRoutes(api, h)

	return engine
}

func RegisterRoutes(api *gin.RouterGroup, h *Handler) {
	books := api.Group("/books")
	{
		books.POST("", h.CreateBook)
		books.GET("", h.ListBooks)
		books.GET("/highest-price", h.GetHighestPriceBook)
		books.GET("/most-purchased", h.GetMostPurchasedBook)
		books.GET("/:id", h.GetBook)
		books.PUT("/:id", h.UpdateBook)
		books.DELETE("/:id", h.DeleteBook)
		books.POST("/:id/purchase", h.PurchaseBook)
		books.POST("/:id/read", h.ReadBook)
	}

	authors := api.Group("/authors")
	{
		authors.POST("", h.CreateAuthor)
		authors.GET("", h.ListAuthors)
		authors.GET("/most-read", h.GetMostReadAuthor)
		authors.GET("/most-purchased", h.GetMostPurchasedAuthor)
		authors.GET("/:id", h.GetAuthor)
		authors.PUT("/:id", h.UpdateAuthor)
		authors.DELETE("/:id", h.DeleteAuthor)
	}

	publishers := api.Group("/publishers")
	{
		publishers.POST("", h.CreatePublisher)
		publishers.GET("", h.ListPublishers)
		publishers.GET("/:id", h.GetPublisher)
		publishers.PUT("/:id", h.UpdatePublisher)
		publishers.DELETE("/:id", h.DeletePublisher)
	}

	categories := api.Group("/categories")
	{
		categories.POST("", h.CreateCategory)
		categories.GET("", h.ListCategories)
		categories.GET("/price-stats", h.GetCategoryPriceStats)
		categories.GET("/:id", h.GetCategory)
		categories.PUT("/:id", h.UpdateCategory)
		categories.DELETE("/:id", h.DeleteCategory)
	}
}
