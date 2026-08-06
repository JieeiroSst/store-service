package http

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), CORSMiddleware())

	engine.GET("/health", h.GetHealth)

	api := engine.Group("/api/v1")
	RegisterRoutes(api, h)

	return engine
}

func RegisterRoutes(api *gin.RouterGroup, h *Handler) {
	notifications := api.Group("/notifications")
	{
		notifications.POST("", h.CreateNotification)
		notifications.POST("/email", h.SendEmail)
		notifications.POST("/slack", h.SendSlack)
		notifications.GET("", h.ListNotifications)
		notifications.GET("/:id", h.GetNotification)
		notifications.PUT("/:id", h.UpdateNotification)
		notifications.DELETE("/:id", h.DeleteNotification)
	}

	devices := api.Group("/devices")
	{
		devices.POST("", h.RegisterDevice)
		devices.GET("", h.ListDevices)
		devices.GET("/:id", h.GetDevice)
		devices.PUT("/:id", h.UpdateDevice)
		devices.DELETE("/:id", h.DeleteDevice)
	}
}
