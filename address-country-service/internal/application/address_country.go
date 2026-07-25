package application

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
)

type addressCountryService struct {
	repo port.AddressCountryRepository
}

func NewAddressCountryService(repo port.AddressCountryRepository) port.AddressCountryUsecase {
	return &addressCountryService{repo: repo}
}

func (s *addressCountryService) Create(ctx context.Context, address *model.AddressCountry) error {
	return s.repo.Create(ctx, address)
}

func (s *addressCountryService) Get(ctx context.Context, id string) (*model.AddressCountry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *addressCountryService) Update(ctx context.Context, address *model.AddressCountry) error {
	return s.repo.Update(ctx, address)
}

func (s *addressCountryService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *addressCountryService) List(ctx context.Context) ([]model.AddressCountry, error) {
	return s.repo.List(ctx)
}
