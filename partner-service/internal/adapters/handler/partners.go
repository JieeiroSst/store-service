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
