package application

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
)

type shippingWeightbasedCountryService struct {
	repo port.ShippingWeightbasedCountryRepository
}

func NewShippingWeightbasedCountryService(
	repo port.ShippingWeightbasedCountryRepository,
) port.ShippingWeightbasedCountryUsecase {
	return &shippingWeightbasedCountryService{repo: repo}
}

func (s *shippingWeightbasedCountryService) Create(ctx context.Context, entry *model.ShippingWeightbasedCountry) error {
	return s.repo.Create(ctx, entry)
}

func (s *shippingWeightbasedCountryService) Get(ctx context.Context, id int) (*model.ShippingWeightbasedCountry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *shippingWeightbasedCountryService) Update(ctx context.Context, entry *model.ShippingWeightbasedCountry) error {
	return s.repo.Update(ctx, entry)
}

func (s *shippingWeightbasedCountryService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *shippingWeightbasedCountryService) List(ctx context.Context) ([]model.ShippingWeightbasedCountry, error) {
	return s.repo.List(ctx)
}
