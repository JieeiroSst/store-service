package usecase

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/dto"
	"github.com/JIeeiroSst/kitchen-service/internal/repository"
	"github.com/JieeiroSst/logger"
)

type Foods interface {
	Create(ctx context.Context, food dto.Food) error
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
}

type FoodUsecase struct {
	FoodRepository repository.Foods
}

func NewFoodUsecase(FoodRepository repository.Foods) *FoodUsecase {
	return &FoodUsecase{
		FoodRepository: FoodRepository,
	}
}

func (u *FoodUsecase) Create(ctx context.Context, food dto.Food) error {
	model := dto.BuildCreateFood(food)

	if err := u.FoodRepository.Create(ctx, model); err != nil {
		return err
	}
	return nil
}

func (u *FoodUsecase) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	foods, err := u.FoodRepository.Find(ctx, pagination)
	if err != nil {
		return logger.Pagination{}, nil
	}
	return foods, nil
}
