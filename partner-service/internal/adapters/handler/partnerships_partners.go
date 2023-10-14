package handler

import (
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

}

func (h *PartnershipsPartnerHandler) ReadPartnershipsPartner(c *gin.Context) {

}

func (h *PartnershipsPartnerHandler) ReadPartnershipsPartners(c *gin.Context) {

}

func (h *PartnershipsPartnerHandler) UpdatePartnershipsPartner(c *gin.Context) {

}

func (h *PartnershipsPartnerHandler) DeletePartnershipsPartner(c *gin.Context) {

}
