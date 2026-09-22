package handler

import (
	"strconv"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/gin-gonic/gin"
)

type PartnerHandler struct {
	svc services.PartnerService
}

func NewPartnerHandler(svc services.PartnerService) *PartnerHandler {
	return &PartnerHandler{
		svc: svc,
	}
}

func (h *PartnerHandler) CreatePartner(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		c.JSON(400, "")
		return
	}
	var partner domain.Partner
	if err := c.ShouldBindJSON(&partner); err != nil {
		c.JSON(400, "")
		return
	}

	if err := h.svc.CreatePartner(userId, partner); err != nil {
		c.JSON(500, err)
		return
	}

	c.JSON(200, "create success")
}

func (h *PartnerHandler) ReadPartner(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}
	partner, err := h.svc.ReadPartner(id)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partner)
}

func (h *PartnerHandler) ReadPartners(c *gin.Context) {
	limmit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))
	pagination := domain.Pagination{
		Limit: limmit,
		Page:  page,
		Sort:  c.Query("sort"),
	}
	partners, err := h.svc.ReadPartners(pagination)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partners)
}

func (h *PartnerHandler) SearchPartners(c *gin.Context) {
	limmit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))
	pagination := domain.Pagination{
		Limit: limmit,
		Page:  page,
		Sort:  c.Query("sort"),
	}

	var filter domain.PartnerFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(400, "")
		return
	}

	partners, err := h.svc.SearchPartners(filter, pagination)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partners)
}

func (h *PartnerHandler) UpdatePartnerStatus(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, "")
		return
	}

	switch body.Status {
	case domain.PartnerStatusActive, domain.PartnerStatusInactive, domain.PartnerStatusPending:
	default:
		c.JSON(400, "invalid status")
		return
	}

	if err := h.svc.UpdatePartnerStatus(id, body.Status); err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, "update status success")
}

func (h *PartnerHandler) UpdatePartner(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}
	var partner domain.Partner
	if err := c.ShouldBindJSON(&partner); err != nil {
		c.JSON(400, "")
		return
	}
	if err := h.svc.UpdatePartner(id, partner); err != nil {
		c.JSON(500, err)
		return
	}

	c.JSON(200, "update success")
}

func (h *PartnerHandler) DeletePartner(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	if err := h.svc.DeletePartner(id); err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, "delete success")
}

func (h *PartnerHandler) RestorePartner(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	if err := h.svc.RestorePartner(id); err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, "restore success")
}

func (h *PartnerHandler) GetPartnerStatistics(c *gin.Context) {
	stats, err := h.svc.GetPartnerStatistics()
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, stats)
}

func (h *PartnerHandler) CheckPartnerActivity(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	activity, err := h.svc.CheckPartnerActivity(id)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, activity)
}

func (h *PartnerHandler) CloseInactivePartners(c *gin.Context) {
	closed, err := h.svc.CloseInactivePartners()
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, gin.H{"closed": closed})
}

func (h *PartnerHandler) AdjustPartnerScore(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	var body struct {
		Points int `json:"points"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, "")
		return
	}

	partner, err := h.svc.AdjustPartnerScore(id, body.Points)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, partner)
}
