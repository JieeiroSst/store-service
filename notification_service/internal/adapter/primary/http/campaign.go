package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createCampaignRequest struct {
	Channel      model.CampaignChannel  `json:"channel"`
	Audience     model.CampaignAudience `json:"audience"`
	Title        string                 `json:"title"`
	Message      string                 `json:"message"`
	Data         map[string]string      `json:"data"`
	TemplateType string                 `json:"template_type"`
	TemplateData map[string]string      `json:"template_data"`
	Topic        string                 `json:"topic"`
	UserIDs      []uint                 `json:"user_ids"`
	Emails       []string               `json:"emails"`
}

func (h *Handler) CreateCampaign(c *gin.Context) {
	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	campaign, err := h.campaign.Create(c.Request.Context(), port.CreateCampaignInput{
		Channel: req.Channel, Audience: req.Audience, Title: req.Title, Message: req.Message, Data: req.Data,
		TemplateType: req.TemplateType, TemplateData: req.TemplateData, Topic: req.Topic,
		UserIDs: req.UserIDs, Emails: req.Emails, RequestedBy: requestedBy(c),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, campaign)
}

func campaignID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

func (h *Handler) GetCampaign(c *gin.Context) {
	id, ok := campaignID(c)
	if !ok {
		return
	}
	view, err := h.campaign.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	items, err := h.campaign.List(c.Request.Context(), limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CancelCampaign(c *gin.Context) {
	id, ok := campaignID(c)
	if !ok {
		return
	}
	view, err := h.campaign.Cancel(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}
