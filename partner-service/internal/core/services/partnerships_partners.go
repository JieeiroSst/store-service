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

func (u *PartnershipsPartnerService) CreatePartnershipsPartner(userID string, PartnershipsPartner domain.PartnershipsPartner) error

func (u *PartnershipsPartnerService) ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error)

func (u *PartnershipsPartnerService) ReadPartnershipsPartners() ([]*domain.PartnershipsPartner, error)

func (u *PartnershipsPartnerService) UpdatePartnershipsPartner(id string, PartnershipsPartner domain.PartnershipsPartner) error

func (u *PartnershipsPartnerService) DeletePartnershipsPartner(id string) error
