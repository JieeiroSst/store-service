package usecase

import (
	"context"

	"github.com/JIeeiroSst/kitchen-service/internal/dto"
	"github.com/JIeeiroSst/kitchen-service/internal/repository"
)

type Categories interface {
	Create(ctx context.Context, category dto.Category) error
	Find(ctx context.Context) ([]dto.Category, error)
}

type CategoryUsecase struct {
	CategoryRepository repository.Categories
}

func NewCategoryUsecase(CategoryRepository repository.Categories) *CategoryUsecase {
	return &CategoryUsecase{
		CategoryRepository: CategoryRepository,
	}
}

func (u *CategoryUsecase) Create(ctx context.Context, category dto.Category) error {
	model := dto.BuildCreateCategory(category)
	if err := u.CategoryRepository.Create(ctx, model); err != nil {
		return err
	}
	return nil
}

func (u *CategoryUsecase) Find(ctx context.Context) ([]dto.Category, error) {
	category, err := u.CategoryRepository.Find(ctx)
	if err != nil {
		return nil, err
	}

	return dto.BuildDtoCategories(category), nil
}
