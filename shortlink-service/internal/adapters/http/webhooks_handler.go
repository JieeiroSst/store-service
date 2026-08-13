package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type WebhooksHandler struct {
	webhooks *app.WebhookService
}

func NewWebhooksHandler(webhooks *app.WebhookService) *WebhooksHandler {
	return &WebhooksHandler{webhooks}
}

func webhookListItemJSON(w *domain.Webhook) gin.H {
	return gin.H{
		"id": w.ID, "user_id": w.UserID, "name": w.Name, "url": w.URL, "events": w.Events,
		"is_active": w.IsActive, "retry_count": w.RetryCount, "timeout_ms": w.TimeoutMs,
		"headers": w.Headers, "created_at": w.CreatedAt, "updated_at": w.UpdatedAt,
	}
}

// webhookFullJSON mirrors `SELECT *` -- includes the secret (get single /
// create / update all return the full row upstream).
func webhookFullJSON(w *domain.Webhook) gin.H {
	m := webhookListItemJSON(w)
	m["secret"] = w.Secret
	return m
}

func (h *WebhooksHandler) List(c *gin.Context) {
	webhooks, err := h.webhooks.List(c.Request.Context(), optionalUserID(c))
	if err != nil {
		respondInternalError(c, "Failed to list webhooks", err)
		return
	}
	out := make([]gin.H, len(webhooks))
	for i, w := range webhooks {
		out[i] = webhookListItemJSON(w)
	}
	c.JSON(http.StatusOK, out)
}

func (h *WebhooksHandler) Get(c *gin.Context) {
	w, err := h.webhooks.Get(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrWebhookNotFound) {
			respondNotFound(c, "Webhook not found")
			return
		}
		respondInternalError(c, "Failed to get webhook", err)
		return
	}
	c.JSON(http.StatusOK, webhookFullJSON(w))
}

type createWebhookRequest struct {
	UserID     *string           `json:"userId" binding:"omitempty,uuid"`
	Name       string            `json:"name" binding:"required,min=1,max=255"`
	URL        string            `json:"url" binding:"required,url"`
	Events     []string          `json:"events" binding:"required,min=1"`
	Headers    map[string]string `json:"headers"`
	RetryCount *int              `json:"retryCount" binding:"omitempty,min=1,max=10"`
	TimeoutMs  *int              `json:"timeoutMs" binding:"omitempty,min=1000,max=60000"`
}

func toWebhookEvents(events []string) []domain.WebhookEvent {
	out := make([]domain.WebhookEvent, len(events))
	for i, e := range events {
		out[i] = domain.WebhookEvent(e)
	}
	return out
}

func (h *WebhooksHandler) Create(c *gin.Context) {
	var req createWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	w, err := h.webhooks.Create(c.Request.Context(), app.CreateWebhookInput{
		UserID: req.UserID, Name: req.Name, URL: req.URL, Events: toWebhookEvents(req.Events),
		Headers: req.Headers, RetryCount: derefInt(req.RetryCount), TimeoutMs: derefInt(req.TimeoutMs),
	})
	if err != nil {
		respondInternalError(c, "Failed to create webhook", err)
		return
	}
	c.JSON(http.StatusOK, webhookFullJSON(w))
}

type updateWebhookRequest struct {
	Name       *string           `json:"name" binding:"omitempty,min=1,max=255"`
	URL        *string           `json:"url" binding:"omitempty,url"`
	Events     []string          `json:"events" binding:"omitempty,min=1"`
	IsActive   *bool             `json:"isActive"`
	Headers    map[string]string `json:"headers"`
	RetryCount *int              `json:"retryCount" binding:"omitempty,min=1,max=10"`
	TimeoutMs  *int              `json:"timeoutMs" binding:"omitempty,min=1000,max=60000"`
}

func (h *WebhooksHandler) Update(c *gin.Context) {
	var req updateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	in := app.UpdateWebhookInput{
		Name: req.Name, URL: req.URL, IsActive: req.IsActive, Headers: req.Headers,
		RetryCount: req.RetryCount, TimeoutMs: req.TimeoutMs,
	}
	if req.Events != nil {
		in.Events = toWebhookEvents(req.Events)
	}

	w, err := h.webhooks.Update(c.Request.Context(), c.Param("id"), optionalUserID(c), in)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrWebhookNotFound):
			respondNotFound(c, "Webhook not found")
		case errors.Is(err, app.ErrNoFieldsToUpdate):
			respondValidationError(c, err)
		default:
			respondInternalError(c, "Failed to update webhook", err)
		}
		return
	}
	c.JSON(http.StatusOK, webhookFullJSON(w))
}

func (h *WebhooksHandler) Delete(c *gin.Context) {
	err := h.webhooks.Delete(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrWebhookNotFound) {
			respondNotFound(c, "Webhook not found")
			return
		}
		respondInternalError(c, "Failed to delete webhook", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *WebhooksHandler) Test(c *gin.Context) {
	result, err := h.webhooks.Test(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrWebhookNotFound) {
			respondNotFound(c, "Webhook not found")
			return
		}
		respondInternalError(c, "Failed to test webhook", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": result.Success, "statusCode": result.StatusCode,
		"responseBody": result.ResponseBody, "error": result.Error,
	})
}
