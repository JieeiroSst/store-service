package application

import (
	"context"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
)

type categoryService struct {
	categories port.CategoryRepository
	ids        port.IDGenerator
}

func NewCategoryService(categories port.CategoryRepository, ids port.IDGenerator) port.CategoryUsecase {
	return &categoryService{categories: categories, ids: ids}
}

func (s *categoryService) CreateCategory(ctx context.Context, input model.CreateCategoryInput) (*model.Category, error) {
	category := &model.Category{
		ID:          s.ids.NewID(),
		Name:        input.Name,
		Description: input.Description,
	}
	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, id string, input model.UpdateCategoryInput) error {
	category := &model.Category{Name: input.Name, Description: input.Description}
	return s.categories.Update(ctx, id, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	return s.categories.Delete(ctx, id)
}

func (s *categoryService) GetCategory(ctx context.Context, id string) (*model.Category, error) {
	return s.categories.GetByID(ctx, id)
}

func (s *categoryService) ListCategories(ctx context.Context, cursor string, limit int) ([]model.Category, string, error) {
	return s.categories.List(ctx, cursor, limit)
}
