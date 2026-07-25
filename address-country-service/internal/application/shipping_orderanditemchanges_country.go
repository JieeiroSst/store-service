package application

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
)

type shippingOrderanditemchangesCountryService struct {
	repo port.ShippingOrderanditemchangesCountryRepository
}

func NewShippingOrderanditemchangesCountryService(
	repo port.ShippingOrderanditemchangesCountryRepository,
) port.ShippingOrderanditemchangesCountryUsecase {
	return &shippingOrderanditemchangesCountryService{repo: repo}
}

func (s *shippingOrderanditemchangesCountryService) Create(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error {
	return s.repo.Create(ctx, entry)
}

func (s *shippingOrderanditemchangesCountryService) Get(ctx context.Context, id int) (*model.ShippingOrderanditemchangesCountry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *shippingOrderanditemchangesCountryService) Update(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error {
	return s.repo.Update(ctx, entry)
}

func (s *shippingOrderanditemchangesCountryService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *shippingOrderanditemchangesCountryService) List(ctx context.Context) ([]model.ShippingOrderanditemchangesCountry, error) {
	return s.repo.List(ctx)
}
