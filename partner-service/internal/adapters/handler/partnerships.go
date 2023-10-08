package handler

import "github.com/JIeeiroSst/partner-service/internal/core/services"

type PartnershipHandler struct {
	svc services.PartnershipService
}

func NewPartnershipHandler(svc services.PartnershipService) *PartnershipHandler {
	return &PartnershipHandler{
		svc: svc,
	}
}
