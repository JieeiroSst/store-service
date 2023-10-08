package handler

import "github.com/JIeeiroSst/partner-service/internal/core/services"

type PartnerHandler struct {
	svc services.PartnerService
}

func NewPartnerHandler(svc services.PartnerService) *PartnerHandler {
	return &PartnerHandler{
		svc: svc,
	}
}
