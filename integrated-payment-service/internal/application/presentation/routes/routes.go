package routes

import (
	"github.com/JIeeiroSst/integrated-payment-service/internal/application/presentation/handlers"
	"github.com/JIeeiroSst/integrated-payment-service/internal/application/presentation/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, paymentHandler *handlers.PaymentHandler) {
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		payments := api.Group("/payments")
		{
			payments.POST("/", paymentHandler.CreatePayment)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.POST("/:id/process", paymentHandler.ProcessPayment)
			payments.POST("/:id/refund", paymentHandler.RefundPayment)
			payments.GET("/:id/status", paymentHandler.GetPaymentStatus)
		}

		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/momo", paymentHandler.HandleMoMoWebhook)
			webhooks.POST("/vnpay", paymentHandler.HandleVNPayWebhook)
			webhooks.POST("/zalopay", paymentHandler.HandleZaloPayWebhook)
			webhooks.POST("/stripe", paymentHandler.HandleStripeWebhook)
		}
	}
}
