package http

import (
	"github.com/JIeeiroSst/doordash-service/pkg/authn"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	api := engine.Group("/api/v1", authenticate(authn.FromEnv()), requireUser())
	{
		orders := api.Group("/orders")
		{
			orders.POST("", forceOwnerBody("customer_id", false), h.CreateOrder)
			orders.GET("", listOrdersAccess, h.ListOrders)
			orders.GET("/:id", h.orderAccess(true), h.GetOrder)
			orders.PATCH("/:id/status", requireAdmin(), h.UpdateOrderStatus)
			orders.POST("/:id/cancel", h.orderAccess(false), h.CancelOrder)
			orders.POST("/:id/assign-driver", requireAdmin(), h.AssignDriver)

			orders.GET("/:id/tracking", h.orderAccess(true), h.ListTracking)
			orders.POST("/:id/tracking", h.orderDriverOrAdmin(), h.AddTrackingEvent)
		}

		assignments := api.Group("/driver-assignments")
		{
			assignments.GET("", ownerQuery("driver_id"), h.ListAssignmentsByDriver)
			assignments.POST("/:id/accept", h.driverOnly(), h.AcceptAssignment)
			assignments.POST("/:id/reject", h.driverOnly(), h.RejectAssignment)
			assignments.POST("/:id/complete", h.driverOnly(), h.CompleteAssignment)
		}
	}

	return engine
}
