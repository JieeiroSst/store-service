package application

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/domain/model"
	"github.com/JIeeiroSst/kitchen-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
)

type categoryService struct {
	repo port.CategoryRepository
}

func NewCategoryService(repo port.CategoryRepository) port.CategoryUsecase {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, category *model.Category) error {
	category.ID = logger.GearedIntID()

	return s.repo.Create(ctx, *category)
}

func (s *categoryService) Find(ctx context.Context) ([]model.Category, error) {
	return s.repo.Find(ctx)
}
