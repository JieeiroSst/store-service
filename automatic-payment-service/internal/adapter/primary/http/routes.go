package http

import "gofr.dev/pkg/gofr"

func RegisterRoutes(app *gofr.App, h *Handler) {
	app.GET("/health", h.GetHealth)

	app.POST("/subscriptions", admin(h.CreateSubscription))
	app.GET("/subscriptions/{id}", admin(h.GetSubscription))
	app.POST("/subscriptions/{id}/cancel", admin(h.CancelSubscription))
	app.GET("/subscriptions/{id}/transactions", admin(h.ListSubscriptionTransactions))
	app.GET("/subscriptions/{id}/invoices", admin(h.ListSubscriptionInvoices))
	app.POST("/subscriptions/renewals/process", admin(h.ProcessRenewals))

	app.POST("/payment-methods", admin(h.AddPaymentMethod))
	app.GET("/payment-methods", admin(h.ListPaymentMethods))
	app.DELETE("/payment-methods/{id}", admin(h.DeletePaymentMethod))
	app.POST("/payment-methods/{id}/default", admin(h.SetDefaultPaymentMethod))

	app.GET("/invoices", admin(h.ListInvoices))
	app.GET("/invoices/{id}", admin(h.GetInvoice))
}
