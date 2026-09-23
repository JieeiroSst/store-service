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

	api := engine.Group("/api/v1")
	{
		payments := api.Group("/payments")
		{
			payments.POST("", h.CreatePayment)
			payments.GET("", h.ListPayments)
			payments.GET("/:id", h.GetPayment)
			payments.POST("/:id/refund", h.RefundPayment)
			payments.GET("/:id/transactions", h.ListTransactions)
		}

		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/:provider", h.HandleWebhook)
		}
	}

	return engine
}
