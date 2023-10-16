package handler

import (
	"strconv"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/gin-gonic/gin"
)

type PartnershipsPartnerHandler struct {
	svc services.PartnershipsPartnerService
}

func NewPartnershipsPartnerHandler(svc services.PartnershipsPartnerService) *PartnershipsPartnerHandler {
	return &PartnershipsPartnerHandler{
		svc: svc,
	}
}

func (h *PartnershipsPartnerHandler) CreatePartnershipsPartner(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		c.JSON(400, "")
		return
	}
	var partnershipsPartner domain.PartnershipsPartner
	if err := c.ShouldBindJSON(&partnershipsPartner); err != nil {
		c.JSON(400, err)
		return
	}
	if err := h.svc.CreatePartnershipsPartner(userId, partnershipsPartner); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, "create success")
}

func (h *PartnershipsPartnerHandler) ReadPartnershipsPartner(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	partnershipsPartner, err := h.svc.ReadPartnershipsPartner(id)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partnershipsPartner)
}

func (h *PartnershipsPartnerHandler) ReadPartnershipsPartners(c *gin.Context) {
	limmit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))
	pagination := domain.Pagination{
		Limit: limmit,
		Page:  page,
		Sort:  c.Query("sort"),
	}
	partnershipsPartners, err := h.svc.ReadPartnershipsPartners(pagination)
	if err != nil {
		c.JSON(500, err)
		return
	}

	c.JSON(200, partnershipsPartners)
}

func (h *PartnershipsPartnerHandler) UpdatePartnershipsPartner(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}
	var partnershipsPartner domain.PartnershipsPartner
	if err := c.ShouldBindJSON(&partnershipsPartner); err != nil {
		c.JSON(400, err)
		return
	}

	if err := h.svc.UpdatePartnershipsPartner(id, partnershipsPartner); err != nil {
		if err != nil {
			c.JSON(500, err)
			return
		}
	}
	c.JSON(200, "update success")
}

func (h *PartnershipsPartnerHandler) DeletePartnershipsPartner(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	if err := h.svc.DeletePartnershipsPartner(id); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, "delete success")
}
