package http

import "gofr.dev/pkg/gofr"

func RegisterRoutes(app *gofr.App, h *Handler) {
	app.POST("/addresses", h.CreateAddress)
	app.GET("/addresses", h.ListAddresses)
	app.GET("/addresses/{id}", h.GetAddress)
	app.PUT("/addresses/{id}", h.UpdateAddress)
	app.DELETE("/addresses/{id}", h.DeleteAddress)

	app.POST("/customers", h.CreateCustomer)
	app.GET("/customers", h.ListCustomers)
	app.GET("/customers/{id}", h.GetCustomer)
	app.PUT("/customers/{id}", h.UpdateCustomer)
	app.DELETE("/customers/{id}", h.DeleteCustomer)

	app.POST("/plans", h.CreatePlan)
	app.GET("/plans", h.ListPlans)
	app.GET("/plans/{id}", h.GetPlan)
	app.PUT("/plans/{id}", h.UpdatePlan)
	app.DELETE("/plans/{id}", h.DeletePlan)

	app.POST("/subscriptions", h.CreateSubscription)
	app.GET("/subscriptions", h.ListSubscriptions)
	app.GET("/subscriptions/{id}", h.GetSubscription)
	app.PUT("/subscriptions/{id}", h.UpdateSubscription)
	app.DELETE("/subscriptions/{id}", h.DeleteSubscription)

	app.POST("/invoices", h.CreateInvoice)
	app.GET("/invoices", h.ListInvoices)
	app.GET("/invoices/{id}", h.GetInvoice)
	app.PUT("/invoices/{id}", h.UpdateInvoice)
	app.DELETE("/invoices/{id}", h.DeleteInvoice)

	app.POST("/transactions", h.CreateTransaction)
	app.GET("/transactions", h.ListTransactions)
	app.GET("/transactions/{id}", h.GetTransaction)
	app.PUT("/transactions/{id}", h.UpdateTransaction)
	app.DELETE("/transactions/{id}", h.DeleteTransaction)
}
