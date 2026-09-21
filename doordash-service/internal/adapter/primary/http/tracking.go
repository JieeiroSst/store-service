package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type addTrackingEventRequest struct {
	Status    string   `json:"status" binding:"required"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Notes     string   `json:"notes"`
}

func (h *Handler) AddTrackingEvent(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req addTrackingEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.tracking.AddTrackingEvent(c.Request.Context(), port.AddTrackingEventInput{
		OrderID:   orderID,
		Status:    model.OrderStatus(req.Status),
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Notes:     req.Notes,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *Handler) ListTracking(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	events, err := h.tracking.ListTrackingByOrder(c.Request.Context(), orderID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, events)
}
