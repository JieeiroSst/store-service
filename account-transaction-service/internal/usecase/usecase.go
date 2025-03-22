package usecase

import "github.com/Jieeirosst/account-transaction-service/internal/repository"

type Usecase struct {
}

type Dependency struct {
	Repos *repository.Repositories
}

func NewUsecase(deps Dependency) *Usecase {

	return &Usecase{}
}
