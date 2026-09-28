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
	notification.RequestedBy = requestedBy(c)
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

type sendEmailRequest struct {
	UserID       uint              `json:"user_id"`
	Recipient    string            `json:"recipient" binding:"required,email"`
	TemplateType string            `json:"template_type"`
	TemplateData map[string]string `json:"template_data"`
	Subject      string            `json:"subject"`
	HTML         string            `json:"html"`
	Text         string            `json:"text"`
	RawData      map[string]any    `json:"raw_data"`
	Priority     int               `json:"priority"`
}

// SendEmail queues a templated email: it saves a pending notification and
// publishes it to RabbitMQ; the worker consumer picks it up, renders
// TemplateType/TemplateData into a subject + HTML body, and sends it via
// the configured EmailSender (Resend).
func (h *Handler) SendEmail(c *gin.Context) {
	var req sendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TemplateType != "" && (req.Subject != "" || req.HTML != "" || req.Text != "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "use either template_type/template_data or subject/html/text, not both"})
		return
	}
	body := req.HTML
	if body == "" && req.Text != "" {
		body = model.TextToHTML(req.Text)
	}
	notification := &model.Notification{
		UserID:       req.UserID,
		Recipient:    req.Recipient,
		Type:         "email",
		Title:        req.Subject,
		Message:      body,
		TemplateType: req.TemplateType,
		TemplateData: req.TemplateData,
		RawData:      req.RawData,
		Priority:     req.Priority,
	}

	notification.RequestedBy = requestedBy(c)
	result, err := h.notification.CreateNotification(c.Request.Context(), notification)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, result)
}

type sendSlackRequest struct {
	UserID       uint              `json:"user_id"`
	TemplateType string            `json:"template_type"`
	TemplateData map[string]string `json:"template_data"`
	Title        string            `json:"title"`
	Text         string            `json:"text"`
	RawData      map[string]any    `json:"raw_data"`
	Priority     int               `json:"priority"`
}

// SendSlack queues a templated Slack message: it saves a pending
// notification and publishes it to RabbitMQ; the worker consumer picks it
// up, renders TemplateType/TemplateData into a title + mrkdwn text, and
// sends it via the configured SlackSender (webhook).
func (h *Handler) SendSlack(c *gin.Context) {
	var req sendSlackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TemplateType != "" && (req.Title != "" || req.Text != "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "use either template_type/template_data or title/text, not both"})
		return
	}
	notification := &model.Notification{
		UserID:       req.UserID,
		Type:         "slack",
		Title:        req.Title,
		Message:      req.Text,
		TemplateType: req.TemplateType,
		TemplateData: req.TemplateData,
		RawData:      req.RawData,
		Priority:     req.Priority,
	}

	notification.RequestedBy = requestedBy(c)
	result, err := h.notification.CreateNotification(c.Request.Context(), notification)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, result)
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
