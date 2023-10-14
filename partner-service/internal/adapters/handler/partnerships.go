package handler

import (
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

}

func (h *PartnershipHandler) ReadPartnership(c *gin.Context) {

}

func (h *PartnershipHandler) ReadPartnerships(c *gin.Context) {

}

func (h *PartnershipHandler) UpdatePartnership(c *gin.Context) {

}

func (h *PartnershipHandler) DeletePartnership(c *gin.Context) {

}
