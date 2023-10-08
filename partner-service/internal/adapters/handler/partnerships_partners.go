package handler

import "github.com/JIeeiroSst/partner-service/internal/core/services"

type PartnershipsPartnerHandler struct {
	svc services.PartnershipsPartnerService
}

func NewPartnershipsPartnerHandler(svc services.PartnershipsPartnerService) *PartnershipsPartnerHandler {
	return &PartnershipsPartnerHandler{
		svc: svc,
	}
}
