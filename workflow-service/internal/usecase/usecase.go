package usecase

import "go.temporal.io/sdk/client"

type Dependency struct {
	Temporal client.Client
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
