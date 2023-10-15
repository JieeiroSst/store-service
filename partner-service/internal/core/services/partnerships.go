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

func (u *PartnerService) CreatePartner(userID string, Partner domain.Partner) error

func (u *PartnerService) ReadPartner(id string) (*domain.Partner, error)

func (u *PartnerService) ReadPartners(pagination domain.Pagination) (*domain.Pagination, error)

func (u *PartnerService) UpdatePartner(id string, Partner domain.Partner) error

func (u *PartnerService) DeletePartner(id string) error
