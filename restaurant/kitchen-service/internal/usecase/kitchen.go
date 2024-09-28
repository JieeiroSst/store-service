package usecase

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/dto"
	"github.com/JIeeiroSst/kitchen-service/internal/repository"
	"github.com/JieeiroSst/logger"
)

type Kitchens interface {
	Create(ctx context.Context, kitchen dto.Kitchen) error
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type KitchenUsecase struct {
	KitchenRepository repository.Kitchens
}

func NewKitchenUsecase(KitchenRepository repository.Kitchens) *KitchenUsecase {
	return &KitchenUsecase{
		KitchenRepository: KitchenRepository,
	}
}

func (u *KitchenUsecase) Create(ctx context.Context, kitchen dto.Kitchen) error {
	model := dto.BuildCreateKitchen(kitchen)
	if err := u.KitchenRepository.Create(ctx, model); err != nil {
		return err
	}
	return nil
}

func (u *KitchenUsecase) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	kitchens, err := u.KitchenRepository.Find(ctx, pagination)
	if err != nil {
		return logger.Pagination{}, nil
	}
	return kitchens, nil
}
