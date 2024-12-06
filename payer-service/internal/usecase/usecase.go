package usecase

import (
	"github.com/JIeeiroSst/payer-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
	Transactions
}

type Dependency struct {
	Repos       *repository.Repository
	CacheHelper expire.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	transactions := NewTransactionUsecase(deps.Repos)

	return &Usecase{
		Transactions: transactions,
	}
}
