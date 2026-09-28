package http

import (
	"github.com/JIeeiroSst/basket-service/internal/adapter/primary/http/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler, authn middleware.Authenticator) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), CORSMiddleware())

	engine.GET("/health", h.GetHealth)

	api := engine.Group("/api/v1", middleware.RequireAuth(authn))
	RegisterRoutes(api, h)

	return engine
}

func RegisterRoutes(api *gin.RouterGroup, h *Handler) {
	baskets := api.Group("/baskets")
	{
		baskets.POST("", h.CreateBasket)
		baskets.GET("", h.ListBaskets)
		baskets.GET("/:id", h.GetBasket)
		baskets.PUT("/:id", h.UpdateBasket)
		baskets.DELETE("/:id", h.DeleteBasket)
	}

	basketLines := api.Group("/basket-lines")
	{
		basketLines.POST("", h.CreateBasketLine)
		basketLines.GET("", h.ListBasketLines)
		basketLines.GET("/:id", h.GetBasketLine)
		basketLines.PUT("/:id", h.UpdateBasketLine)
		basketLines.DELETE("/:id", h.DeleteBasketLine)
	}

	basketLineAttributes := api.Group("/basket-line-attributes")
	{
		basketLineAttributes.POST("", h.CreateBasketLineAttribute)
		basketLineAttributes.GET("", h.ListBasketLineAttributes)
		basketLineAttributes.GET("/:id", h.GetBasketLineAttribute)
		basketLineAttributes.PUT("/:id", h.UpdateBasketLineAttribute)
		basketLineAttributes.DELETE("/:id", h.DeleteBasketLineAttribute)
	}

	orders := api.Group("/orders")
	{
		orders.GET("", h.ListOrders)
		orders.GET("/:id", h.GetOrder)
	}

	users := api.Group("/users")
	{
		users.GET("/:id", h.GetUser)
	}
}
