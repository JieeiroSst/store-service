package application

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
)

type orderShippingaddressService struct {
	repo port.OrderShippingaddressRepository
}

func NewOrderShippingaddressService(repo port.OrderShippingaddressRepository) port.OrderShippingaddressUsecase {
	return &orderShippingaddressService{repo: repo}
}

func (s *orderShippingaddressService) Create(ctx context.Context, address *model.OrderShippingaddress) error {
	return s.repo.Create(ctx, address)
}

func (s *orderShippingaddressService) Get(ctx context.Context, id int) (*model.OrderShippingaddress, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderShippingaddressService) Update(ctx context.Context, address *model.OrderShippingaddress) error {
	return s.repo.Update(ctx, address)
}

func (s *orderShippingaddressService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *orderShippingaddressService) List(ctx context.Context) ([]model.OrderShippingaddress, error) {
	return s.repo.List(ctx)
}
