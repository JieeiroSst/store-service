package usecase

import "github.com/JIeeiroSst/order-service/internal/repository"

type Usecase struct {
	Orders
}

type Dependency struct {
	Repos *repository.Repository
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Orders: NewOrderUsecase(deps.Repos.Orders),
	}
}
