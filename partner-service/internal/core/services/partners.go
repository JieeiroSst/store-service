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

func (u *PartnerService) SearchPartners(filter domain.PartnerFilter, pagination domain.Pagination) (*domain.Pagination, error) {
	return u.repo.SearchPartners(filter, pagination)
}

func (u *PartnerService) UpdatePartner(id string, Partner domain.Partner) error {
	return u.repo.UpdatePartner(id, Partner)
}

func (u *PartnerService) UpdatePartnerStatus(id string, status string) error {
	return u.repo.UpdatePartnerStatus(id, status)
}

func (u *PartnerService) DeletePartner(id string) error {
	return u.repo.DeletePartner(id)
}

func (u *PartnerService) RestorePartner(id string) error {
	return u.repo.RestorePartner(id)
}

func (u *PartnerService) GetPartnerStatistics() (*domain.PartnerStatistics, error) {
	return u.repo.GetPartnerStatistics()
}

func (u *PartnerService) CheckPartnerActivity(id string) (*domain.PartnerActivity, error) {
	return u.repo.CheckPartnerActivity(id)
}

func (u *PartnerService) CloseInactivePartners() (int64, error) {
	return u.repo.CloseInactivePartners()
}

func (u *PartnerService) AdjustPartnerScore(id string, delta int) (*domain.Partner, error) {
	return u.repo.AdjustPartnerScore(id, delta)
}
