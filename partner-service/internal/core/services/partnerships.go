package services

import (
	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/ports"
)

type PartnershipService struct {
	repo ports.PartnershipRepository
}

func NewPartnershipService(repo ports.PartnershipRepository) *PartnershipService {
	return &PartnershipService{
		repo: repo,
	}
}

func (u *PartnershipService) CreatePartnership(userID string, Partnership domain.Partnership) error {
	return u.repo.CreatePartnership(userID, Partnership)
}

func (u *PartnershipService) ReadPartnership(id string) (*domain.Partnership, error) {
	return u.repo.ReadPartnership(id)
}

func (u *PartnershipService) ReadPartnerships(pagination domain.Pagination) (*domain.Pagination, error) {
	return u.repo.ReadPartnerships(pagination)
}

func (u *PartnershipService) UpdatePartnership(id string, Partnership domain.Partnership) error {
	return u.repo.UpdatePartnership(id, Partnership)
}

func (u *PartnershipService) DeletePartnership(id string) error {
	return u.repo.DeletePartnership(id)
}
