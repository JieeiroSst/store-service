package services

import "github.com/JIeeiroSst/partner-service/internal/core/ports"

type PartnershipsPartnerService struct {
	repo ports.PartnershipsPartnerRepository
}

func NewPartnershipsPartnerService(repo ports.PartnershipsPartnerRepository) *PartnershipsPartnerService {
	return &PartnershipsPartnerService{
		repo: repo,
	}
}
