package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterDevice(c *gin.Context) {
	var device model.UserDevice
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.device.RegisterDevice(c.Request.Context(), &device)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.device.GetDevice(c.Request.Context(), uint(id))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListDevices(c *gin.Context) {
	result, err := h.device.ListDevices(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdateDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var device model.UserDevice
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	device.ID = uint(id)
	result, err := h.device.UpdateDevice(c.Request.Context(), &device)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.device.DeleteDevice(c.Request.Context(), uint(id)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
