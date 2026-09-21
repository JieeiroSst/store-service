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
		orders := api.Group("/orders")
		{
			orders.POST("", h.CreateOrder)
			orders.GET("", h.ListOrders) // ?customer_id= or ?restaurant_id=
			orders.GET("/:id", h.GetOrder)
			orders.PATCH("/:id/status", h.UpdateOrderStatus)
			orders.POST("/:id/cancel", h.CancelOrder)
			orders.POST("/:id/assign-driver", h.AssignDriver)

			orders.GET("/:id/tracking", h.ListTracking)
			orders.POST("/:id/tracking", h.AddTrackingEvent)
		}

		assignments := api.Group("/driver-assignments")
		{
			assignments.GET("", h.ListAssignmentsByDriver) // ?driver_id=
			assignments.POST("/:id/accept", h.AcceptAssignment)
			assignments.POST("/:id/reject", h.RejectAssignment)
			assignments.POST("/:id/complete", h.CompleteAssignment)
		}
	}

	return engine
}
