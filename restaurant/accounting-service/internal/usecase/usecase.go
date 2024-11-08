package usecase

import "github.com/JIeeiroSst/accounting-service/internal/repository"

type Usecase struct {
	AuthCarts
}

type Dependency struct {
	Repos *repository.Repository
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		AuthCarts: NewAuthCartUsecase(deps.Repos.AuthCarts),
	}
}
