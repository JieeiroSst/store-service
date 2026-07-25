package application

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
)

type partneraddressService struct {
	repo port.PartneraddressRepository
}

func NewPartneraddressService(repo port.PartneraddressRepository) port.PartneraddressUsecase {
	return &partneraddressService{repo: repo}
}

func (s *partneraddressService) Create(ctx context.Context, address *model.Partneraddress) error {
	return s.repo.Create(ctx, address)
}

func (s *partneraddressService) Get(ctx context.Context, id int) (*model.Partneraddress, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *partneraddressService) Update(ctx context.Context, address *model.Partneraddress) error {
	return s.repo.Update(ctx, address)
}

func (s *partneraddressService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *partneraddressService) List(ctx context.Context) ([]model.Partneraddress, error) {
	return s.repo.List(ctx)
}
