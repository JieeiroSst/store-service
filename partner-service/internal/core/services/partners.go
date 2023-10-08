package services

import "github.com/JIeeiroSst/partner-service/internal/core/ports"

type PartnershipService struct {
	repo ports.PartnerRepository
}

func NewPartnershipService(repo ports.PartnerRepository) *PartnershipService {
	return &PartnershipService{
		repo: repo,
	}
}
