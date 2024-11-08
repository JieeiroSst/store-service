package usecase

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/internal/dto"
	"github.com/JIeeiroSst/accounting-service/internal/repository"
)

type AuthCarts interface {
	SaveDelivery(ctx context.Context, delivery dto.Delivery) error
	SaveOrder(ctx context.Context, order dto.Order) error
}

type authCartUsecase struct {
	authCartRepository repository.AuthCarts
}

func NewAuthCartUsecase(authCartRepository repository.AuthCarts) *authCartUsecase {
	return &authCartUsecase{
		authCartRepository: authCartRepository,
	}
}

func (u *authCartUsecase) SaveDelivery(ctx context.Context, delivery dto.Delivery) error {
	if err := u.authCartRepository.SaveDelivery(ctx, delivery.Build()); err != nil {
		return err
	}
	return nil
}

func (u *authCartUsecase) SaveOrder(ctx context.Context, order dto.Order) error {
	if err := u.authCartRepository.SaveOrder(ctx, order.Build()); err != nil {
		return err
	}
	return nil
}
