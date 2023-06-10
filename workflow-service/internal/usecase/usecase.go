package usecase

import (
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"go.temporal.io/sdk/client"
)

type Dependency struct {
	Temporal client.Client
	Repository *repository.Repositories
}

type Usecase struct {
	Cards
}

func NewUsecase(deps Dependency) *Usecase {
	cardUsecase := NewCardUsecase(deps.Temporal)
	return &Usecase{
		Cards: cardUsecase,
	}
}
