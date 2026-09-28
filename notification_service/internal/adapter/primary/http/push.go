package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type sendPushRequest struct {
	UserID   uint              `json:"user_id"`
	Email    string            `json:"email"`
	Phone    string            `json:"phone"`
	Title    string            `json:"title"`
	Message  string            `json:"message"`
	Data     map[string]string `json:"data"`
	Priority int               `json:"priority"`
}

func (h *Handler) SendPush(c *gin.Context) {
	var req sendPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.notification.CreateNotification(c.Request.Context(), &model.Notification{
		UserID:      req.UserID,
		Email:       req.Email,
		Phone:       req.Phone,
		Title:       req.Title,
		Message:     req.Message,
		Data:        req.Data,
		Priority:    req.Priority,
		Type:        "push",
		RequestedBy: requestedBy(c),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, result)
}

type contactRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func userIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return 0, false
	}
	return uint(id), true
}

func (h *Handler) UpsertContact(c *gin.Context) {
	id, ok := userIDParam(c)
	if !ok {
		return
	}
	var req contactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contact, err := h.device.UpsertContact(c.Request.Context(), port.ContactInput{UserID: id, Email: req.Email, Phone: req.Phone})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, contact)
}

func (h *Handler) GetContact(c *gin.Context) {
	id, ok := userIDParam(c)
	if !ok {
		return
	}
	contact, err := h.device.GetContact(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, contact)
}
