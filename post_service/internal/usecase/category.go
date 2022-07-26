package usecase

import (
	"github.com/JIeeiroSst/post-service/internal/repository"
	"github.com/JIeeiroSst/post-service/model"
	"github.com/JIeeiroSst/post-service/pkg/snowflake"
)

type Categories interface {
	Create(category model.Category) error
	Update(id string, category model.Category) error
	Delete(id string) error
	Categories() ([]model.Category, error)
	CategoryById(id string) (*model.Category, error)
}

type CategoryUsecase struct {
	CategoryRepo repository.Categories
	Snowflake    snowflake.SnowflakeData
}

func NewCategoryUsecase(CategoryRepo repository.Categories,
	Snowflake snowflake.SnowflakeData) *CategoryUsecase {
	return &CategoryUsecase{
		CategoryRepo: CategoryRepo,
		Snowflake:    Snowflake,
	}
}

func (u *CategoryUsecase) Create(category model.Category) error {
	args := model.Category{
		Id:          u.Snowflake.GearedID(),
		Name:        category.Name,
		Description: category.Description,
	}
	if err := u.CategoryRepo.Create(args); err != nil {
		return err
	}
	return nil
}

func (u *CategoryUsecase) Update(id string, category model.Category) error {
	args := model.Category{
		Name:        category.Name,
		Description: category.Description,
	}
	if err := u.CategoryRepo.Update(id, args); err != nil {
		return err
	}
	return nil
}

func (u *CategoryUsecase) Delete(id string) error {
	if err := u.CategoryRepo.Delete(id); err != nil {
		return err
	}
	return nil
}

func (u *CategoryUsecase) Categories() ([]model.Category, error) {
	categories, err := u.CategoryRepo.Categories()
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (u *CategoryUsecase) CategoryById(id string) (*model.Category, error) {
	category, err := u.CategoryRepo.CategoryById(id)
	if err != nil {
		return nil, err
	}
	return category, nil
}
