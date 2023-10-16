package handler

import (
	"strconv"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/gin-gonic/gin"
)

type PartnershipHandler struct {
	svc services.PartnershipService
}

func NewPartnershipHandler(svc services.PartnershipService) *PartnershipHandler {
	return &PartnershipHandler{
		svc: svc,
	}
}

func (h *PartnershipHandler) CreatePartnership(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		c.JSON(400, "")
		return
	}
	var partnership domain.Partnership
	if err := c.ShouldBindJSON(&partnership); err != nil {
		c.JSON(400, err)
		return
	}
	if err := h.svc.CreatePartnership(userId, partnership); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partnership)
}

func (h *PartnershipHandler) ReadPartnership(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	partnership, err := h.svc.ReadPartnership(id)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partnership)
}

func (h *PartnershipHandler) ReadPartnerships(c *gin.Context) {
	limmit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))
	pagination := domain.Pagination{
		Limit: limmit,
		Page:  page,
		Sort:  c.Query("sort"),
	}
	partnerships, err := h.svc.ReadPartnerships(pagination)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partnerships)
}

func (h *PartnershipHandler) UpdatePartnership(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}
	var partnership domain.Partnership
	if err := c.ShouldBindJSON(&partnership); err != nil {
		c.JSON(400, err)
		return
	}
	if err := h.svc.UpdatePartnership(id, partnership); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, partnership)
}

func (h *PartnershipHandler) DeletePartnership(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}
	if err := h.svc.DeletePartnership(id); err != nil {
		c.JSON(400, err)
		return
	}
	c.JSON(200, "delete success")
}
