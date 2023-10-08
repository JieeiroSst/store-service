package services

import "github.com/JIeeiroSst/partner-service/internal/core/ports"

type PartnerService struct {
	repo ports.PartnerRepository
}

func NewPartnerService(repo ports.PartnerRepository) *PartnerService {
	return &PartnerService{
		repo: repo,
	}
}
