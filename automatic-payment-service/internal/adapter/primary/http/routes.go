package http

import "gofr.dev/pkg/gofr"

func RegisterRoutes(app *gofr.App, h *Handler) {
	app.GET("/health", h.GetHealth)

	app.POST("/subscriptions", h.CreateSubscription)
	app.GET("/subscriptions/{id}", h.GetSubscription)
	app.POST("/subscriptions/{id}/cancel", h.CancelSubscription)
	app.GET("/subscriptions/{id}/transactions", h.ListSubscriptionTransactions)
	app.GET("/subscriptions/{id}/invoices", h.ListSubscriptionInvoices)
	app.POST("/subscriptions/renewals/process", h.ProcessRenewals)

	app.POST("/payment-methods", h.AddPaymentMethod)
	app.GET("/payment-methods", h.ListPaymentMethods)
	app.DELETE("/payment-methods/{id}", h.DeletePaymentMethod)
	app.POST("/payment-methods/{id}/default", h.SetDefaultPaymentMethod)

	app.GET("/invoices", h.ListInvoices)
	app.GET("/invoices/{id}", h.GetInvoice)
}
