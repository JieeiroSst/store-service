package application

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
)

type addressService struct {
	repo port.AddressRepository
}

func NewAddressService(repo port.AddressRepository) port.AddressUsecase {
	return &addressService{repo: repo}
}

func (s *addressService) Create(ctx context.Context, address *model.Address) error {
	return s.repo.Create(ctx, address)
}

func (s *addressService) Get(ctx context.Context, id int) (*model.Address, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *addressService) Update(ctx context.Context, address *model.Address) error {
	return s.repo.Update(ctx, address)
}

func (s *addressService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *addressService) List(ctx context.Context) ([]model.Address, error) {
	return s.repo.List(ctx)
}
