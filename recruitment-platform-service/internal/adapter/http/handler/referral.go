package handler

import (
	"net/http"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReferralHandler struct {
	svc port.ReferralService
}

func NewReferralHandler(svc port.ReferralService) *ReferralHandler {
	return &ReferralHandler{svc: svc}
}

func (h *ReferralHandler) RegisterRoutesProtected(r gin.IRouter) {
	g := r.Group("/referrals")
	g.POST("/partners", h.RegisterPartner)
	g.GET("/partners/:id/stats", h.GetPartnerStats)
	g.GET("/partners/:id/network", h.GetPartnerNetwork)
	g.POST("/partners/:id/payout", h.RequestPayout)
	g.POST("/links", h.GenerateLink)
	g.GET("/leaderboard", h.Leaderboard)
}

func (h *ReferralHandler) TrackClick(c *gin.Context) {
	token := c.Param("token")
	if err := h.svc.TrackReferralClick(c.Request.Context(), token); err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.Redirect(http.StatusFound, "/apply?ref="+token)
}

func (h *ReferralHandler) RegisterPartner(c *gin.Context) {
	var req struct {
		FullName        string `json:"full_name"         binding:"required"`
		Email           string `json:"email"             binding:"required,email"`
		Phone           string `json:"phone"`
		Company         string `json:"company"`
		ReferredByToken string `json:"referred_by_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	partner, err := h.svc.RegisterPartner(c.Request.Context(), port.RegisterPartnerCommand{
		UserID:          mustGetUserID(c),
		FullName:        req.FullName,
		Email:           req.Email,
		Phone:           req.Phone,
		Company:         req.Company,
		ReferredByToken: req.ReferredByToken,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, partner)
}

func (h *ReferralHandler) GetPartnerStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	stats, err := h.svc.GetPartnerStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "partner not found"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *ReferralHandler) GetPartnerNetwork(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	stats, err := h.svc.GetPartnerStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "partner not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"partner_id":      id,
		"tier":            stats.Partner.Tier,
		"network_depth":   stats.Partner.NetworkDepth,
		"total_referrals": stats.TotalReferrals,
		"hired_count":     stats.HiredCount,
	})
}

func (h *ReferralHandler) RequestPayout(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	payout, err := h.svc.RequestPayout(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, payout)
}

func (h *ReferralHandler) GenerateLink(c *gin.Context) {
	var req struct {
		PartnerID string  `json:"partner_id" binding:"required,uuid"`
		JobID     *string `json:"job_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	partnerID := uuid.MustParse(req.PartnerID)
	var jobID *uuid.UUID
	if req.JobID != nil {
		parsed := uuid.MustParse(*req.JobID)
		jobID = &parsed
	}
	ref, err := h.svc.GenerateReferralLink(c.Request.Context(), partnerID, jobID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"referral":   ref,
		"link":       "https://yourdomain.com/apply?ref=" + ref.Token,
		"expires_at": ref.ExpiresAt,
	})
}

func (h *ReferralHandler) Leaderboard(c *gin.Context) {
	stats, err := h.svc.GetLeaderboard(c.Request.Context(), 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"leaderboard": stats})
}
