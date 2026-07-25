package application

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
)

type orderBillingaddressService struct {
	repo port.OrderBillingaddressRepository
}

func NewOrderBillingaddressService(repo port.OrderBillingaddressRepository) port.OrderBillingaddressUsecase {
	return &orderBillingaddressService{repo: repo}
}

func (s *orderBillingaddressService) Create(ctx context.Context, address *model.OrderBillingaddress) error {
	return s.repo.Create(ctx, address)
}

func (s *orderBillingaddressService) Get(ctx context.Context, id int) (*model.OrderBillingaddress, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderBillingaddressService) Update(ctx context.Context, address *model.OrderBillingaddress) error {
	return s.repo.Update(ctx, address)
}

func (s *orderBillingaddressService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *orderBillingaddressService) List(ctx context.Context) ([]model.OrderBillingaddress, error) {
	return s.repo.List(ctx)
}
