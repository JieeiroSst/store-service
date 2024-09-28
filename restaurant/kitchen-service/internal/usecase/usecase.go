package usecase

import "github.com/JIeeiroSst/kitchen-service/internal/repository"

type Usecase struct {
	Categories
	Foods
	Kitchens
}

type Dependency struct {
	Repos *repository.Repository
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Categories: NewCategoryUsecase(deps.Repos.Categories),
		Foods:      NewFoodUsecase(deps.Repos.Foods),
		Kitchens:   NewKitchenUsecase(deps.Repos.Kitchens),
	}
}
