package application

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
)

type userAddressService struct {
	repo port.UserAddressRepository
}

func NewUserAddressService(repo port.UserAddressRepository) port.UserAddressUsecase {
	return &userAddressService{repo: repo}
}

func (s *userAddressService) Create(ctx context.Context, address *model.UserAddress) error {
	return s.repo.Create(ctx, address)
}

func (s *userAddressService) Get(ctx context.Context, id int) (*model.UserAddress, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userAddressService) Update(ctx context.Context, address *model.UserAddress) error {
	return s.repo.Update(ctx, address)
}

func (s *userAddressService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *userAddressService) List(ctx context.Context) ([]model.UserAddress, error) {
	return s.repo.List(ctx)
}
