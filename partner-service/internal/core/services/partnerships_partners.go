package services

import (
	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/ports"
)

type PartnershipsPartnerService struct {
	repo ports.PartnershipsPartnerRepository
}

func NewPartnershipsPartnerService(repo ports.PartnershipsPartnerRepository) *PartnershipsPartnerService {
	return &PartnershipsPartnerService{
		repo: repo,
	}
}

func (u *PartnershipsPartnerService) CreatePartnershipsPartner(userID string, PartnershipsPartner domain.PartnershipsPartner) error {
	return u.repo.CreatePartnershipsPartner(userID, PartnershipsPartner)
}

func (u *PartnershipsPartnerService) ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error) {
	return u.repo.ReadPartnershipsPartner(id)
}

func (u *PartnershipsPartnerService) ReadPartnershipsPartners(pagination domain.Pagination) (*domain.Pagination, error) {
	return u.repo.ReadPartnershipsPartners(pagination)
}

func (u *PartnershipsPartnerService) UpdatePartnershipsPartner(id string, PartnershipsPartner domain.PartnershipsPartner) error {
	return u.repo.UpdatePartnershipsPartner(id, PartnershipsPartner)
}

func (u *PartnershipsPartnerService) DeletePartnershipsPartner(id string) error {
	return u.repo.DeletePartnershipsPartner(id)
}
