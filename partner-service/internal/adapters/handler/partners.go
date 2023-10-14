package handler

import (
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

}

func (h *PartnerHandler) ReadPartner(c *gin.Context) {

}

func (h *PartnerHandler) ReadPartners(c *gin.Context) {

}

func (h *PartnerHandler) UpdatePartner(c *gin.Context) {

}

func (h *PartnerHandler) DeletePartner(c *gin.Context) {

}
