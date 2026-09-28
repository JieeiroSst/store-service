package http

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", h.Health)
	r.GET("/health/ready", h.Ready)

	v1 := r.Group("/api/v1")

	v1.POST("/machines", h.RegisterMachine)
	v1.GET("/machines", h.ListMachines)
	v1.GET("/machines/:id", h.GetMachine)
	v1.PATCH("/machines/:id", h.UpdateMachine)
	v1.PATCH("/machines/:id/status", h.ChangeMachineStatus)
	v1.POST("/machines/:id/maintenance", h.RecordMaintenance)
	v1.GET("/machines/:id/maintenance", h.ListMaintenance)
	v1.GET("/machines/:id/events", h.ListEvents)
	v1.GET("/machines/:id/sales", h.SalesReport)
	v1.GET("/machines/:id/inventory", h.ListInventory)
	v1.PUT("/machines/:id/slots/:slot", h.AssignSlot)

	v1.GET("/inventory/low", h.ListLowInventory)
	v1.POST("/inventory/:id/restock", h.Restock)

	v1.POST("/categories", h.CreateCategory)
	v1.GET("/categories", h.ListCategories)
	v1.POST("/products", h.CreateProduct)
	v1.GET("/products", h.ListProducts)
	v1.GET("/products/:id", h.GetProduct)
	v1.PATCH("/products/:id", h.UpdateProduct)

	v1.POST("/sessions", h.StartSession)
	v1.GET("/sessions/:id", h.GetSession)
	v1.POST("/sessions/:id/end", h.EndSession)
	v1.GET("/sessions/:id/orders", h.ListSessionOrders)
	v1.POST("/sessions/:id/reservations", h.Reserve)
	v1.DELETE("/reservations/:id", h.CancelReservation)
	v1.POST("/reservations/:id/checkout", h.Checkout)

	v1.GET("/payments/:id", h.GetPayment)
	v1.GET("/orders/:id", h.GetOrder)
	v1.POST("/orders/:id/dispense", h.ReportDispense)
	v1.POST("/orders/:id/refund", h.RefundOrder)
	return r
}
