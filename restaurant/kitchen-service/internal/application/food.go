package application

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
)

type foodService struct {
	repo port.FoodRepository
}

func NewFoodService(repo port.FoodRepository) port.FoodUsecase {
	return &foodService{repo: repo}
}

func (s *foodService) Create(ctx context.Context, food *model.Food) error {
	food.ID = logger.GearedIntID()

	return s.repo.Create(ctx, *food)
}

func (s *foodService) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	return s.repo.Find(ctx, pagination)
}

func (s *foodService) FindByIDs(ctx context.Context, ids []int) ([]model.Food, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return s.repo.FindByIDs(ctx, ids)
}
