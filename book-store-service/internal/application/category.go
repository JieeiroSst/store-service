package application

import (
	"context"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
)

type categoryService struct {
	repo port.CategoryRepository
}

func NewCategoryService(repo port.CategoryRepository) port.CategoryUsecase {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *model.Category) (*model.Category, error) {
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) GetCategory(ctx context.Context, id int) (*model.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *categoryService) ListCategories(ctx context.Context) ([]model.Category, error) {
	return s.repo.List(ctx)
}

func (s *categoryService) UpdateCategory(ctx context.Context, category *model.Category) (*model.Category, error) {
	if _, err := s.repo.GetByID(ctx, category.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
