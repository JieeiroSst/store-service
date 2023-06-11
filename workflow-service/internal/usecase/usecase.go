package usecase

import (
	"github.com/JIeeiroSst/workflow-service/internal/activities/card"
	"go.temporal.io/sdk/client"
)

type Dependency struct {
	Temporal client.Client
	Card     card.CardWorkflow
}

type Usecase struct {
	Cards
}

func NewUsecase(deps Dependency) *Usecase {
	cardUsecase := NewCardUsecase(deps.Temporal, deps.Card)
	return &Usecase{
		Cards: cardUsecase,
	}
}
