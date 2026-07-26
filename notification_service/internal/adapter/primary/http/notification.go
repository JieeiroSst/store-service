package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateNotification(c *gin.Context) {
	var notification model.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.notification.CreateNotification(c.Request.Context(), &notification)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.notification.GetNotification(c.Request.Context(), uint(id))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListNotifications(c *gin.Context) {
	result, err := h.notification.ListNotifications(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdateNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var notification model.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	notification.ID = uint(id)
	result, err := h.notification.UpdateNotification(c.Request.Context(), &notification)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.notification.DeleteNotification(c.Request.Context(), uint(id)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
