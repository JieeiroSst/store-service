package services

import (
	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/ports"
)

type PartnerService struct {
	repo ports.PartnerRepository
}

func NewPartnerService(repo ports.PartnerRepository) *PartnerService {
	return &PartnerService{
		repo: repo,
	}
}

func (u *PartnerService) CreatePartner(userID string, Partner domain.Partner) error {
	return u.repo.CreatePartner(userID, Partner)
}

func (u *PartnerService) ReadPartner(id string) (*domain.Partner, error) {
	return u.repo.ReadPartner(id)
}

func (u *PartnerService) ReadPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	return u.repo.ReadPartners(pagination)
}

func (u *PartnerService) UpdatePartner(id string, Partner domain.Partner) error {
	return u.repo.UpdatePartner(id, Partner)
}

func (u *PartnerService) DeletePartner(id string) error {
	return u.repo.DeletePartner(id)
}
