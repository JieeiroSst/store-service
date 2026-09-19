package port

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JieeiroSst/logger"
)

type KitchenUsecase interface {
	Create(ctx context.Context, kitchen *model.Kitchen) error
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	UpdateStatus(ctx context.Context, id int, status string) error
}

type FoodUsecase interface {
	Create(ctx context.Context, food *model.Food) error
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	FindByIDs(ctx context.Context, ids []int) ([]model.Food, error)
}

type CategoryUsecase interface {
	Create(ctx context.Context, category *model.Category) error
	Find(ctx context.Context) ([]model.Category, error)
}
