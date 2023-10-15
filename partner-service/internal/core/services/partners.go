package services

import (
	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/ports"
)

type PartnershipService struct {
	repo ports.PartnerRepository
}

func NewPartnershipService(repo ports.PartnerRepository) *PartnershipService {
	return &PartnershipService{
		repo: repo,
	}
}

func (u *PartnershipService) CreatePartnership(userID string, Partnership domain.Partnership) error

func (u *PartnershipService) ReadPartnership(id string) (*domain.Partnership, error)

func (u *PartnershipService) ReadPartnerships(pagination domain.Pagination) (*domain.Pagination, error)

func (u *PartnershipService) UpdatePartnership(id string, Partnership domain.Partnership) error

func (u *PartnershipService) DeletePartnership(id string) error
