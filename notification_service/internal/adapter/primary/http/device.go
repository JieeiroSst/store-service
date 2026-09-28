package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type deviceView struct {
	model.UserDevice
	TokenPreview string `json:"token_preview"`
}

func viewOf(d *model.UserDevice) deviceView {
	return deviceView{UserDevice: *d, TokenPreview: d.TokenPreview()}
}

func viewsOf(ds []model.UserDevice) []deviceView {
	out := make([]deviceView, len(ds))
	for i := range ds {
		out[i] = viewOf(&ds[i])
	}
	return out
}

type registerDeviceRequest struct {
	UserID      uint   `json:"user_id"`
	DeviceToken string `json:"device_token"`
	DeviceID    string `json:"device_id"`
	DeviceType  string `json:"device_type"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

func (h *Handler) RegisterDevice(c *gin.Context) {
	var req registerDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.device.RegisterDevice(c.Request.Context(), port.RegisterDeviceInput{
		UserID:     req.UserID,
		Token:      req.DeviceToken,
		DeviceID:   req.DeviceID,
		DeviceType: req.DeviceType,
		Email:      req.Email,
		Phone:      req.Phone,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, viewOf(result))
}

type unregisterDeviceRequest struct {
	DeviceToken string `json:"device_token"`
	DeviceID    string `json:"device_id"`
	UserID      uint   `json:"user_id"`
}

func (h *Handler) UnregisterDevice(c *gin.Context) {
	var req unregisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.device.UnregisterDevice(c.Request.Context(), port.UnregisterDeviceInput{Token: req.DeviceToken, DeviceID: req.DeviceID, UserID: req.UserID})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deactivated": n})
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
	c.JSON(http.StatusOK, viewOf(result))
}

func (h *Handler) ListDevices(c *gin.Context) {
	userID, ok := parseUint(c, "user_id")
	if !ok {
		return
	}
	result, err := h.device.ListDevices(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, viewsOf(result))
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
	c.JSON(http.StatusOK, viewOf(result))
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
