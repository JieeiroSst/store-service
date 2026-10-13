package http

import "gofr.dev/pkg/gofr"

func RegisterRoutes(app *gofr.App, h *Handler) {
	app.POST("/addresses", admin(h.CreateAddress))
	app.GET("/addresses", admin(h.ListAddresses))
	app.GET("/addresses/{id}", admin(h.GetAddress))
	app.PUT("/addresses/{id}", admin(h.UpdateAddress))
	app.DELETE("/addresses/{id}", admin(h.DeleteAddress))

	app.POST("/customers", admin(h.CreateCustomer))
	app.GET("/customers", admin(h.ListCustomers))
	app.GET("/customers/{id}", admin(h.GetCustomer))
	app.PUT("/customers/{id}", admin(h.UpdateCustomer))
	app.DELETE("/customers/{id}", admin(h.DeleteCustomer))

	app.POST("/plans", admin(h.CreatePlan))
	app.GET("/plans", h.ListPlans)
	app.GET("/plans/{id}", h.GetPlan)
	app.PUT("/plans/{id}", admin(h.UpdatePlan))
	app.DELETE("/plans/{id}", admin(h.DeletePlan))

	app.POST("/subscriptions", admin(h.CreateSubscription))
	app.GET("/subscriptions", admin(h.ListSubscriptions))
	app.GET("/subscriptions/{id}", admin(h.GetSubscription))
	app.PUT("/subscriptions/{id}", admin(h.UpdateSubscription))
	app.DELETE("/subscriptions/{id}", admin(h.DeleteSubscription))

	app.POST("/invoices", admin(h.CreateInvoice))
	app.GET("/invoices", admin(h.ListInvoices))
	app.GET("/invoices/{id}", admin(h.GetInvoice))
	app.PUT("/invoices/{id}", admin(h.UpdateInvoice))
	app.DELETE("/invoices/{id}", admin(h.DeleteInvoice))

	app.POST("/transactions", admin(h.CreateTransaction))
	app.GET("/transactions", admin(h.ListTransactions))
	app.GET("/transactions/{id}", admin(h.GetTransaction))
	app.PUT("/transactions/{id}", admin(h.UpdateTransaction))
	app.DELETE("/transactions/{id}", admin(h.DeleteTransaction))
}
