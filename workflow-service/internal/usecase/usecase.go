package usecase

import (
	"github.com/JIeeiroSst/workflow-service/internal/activities/card"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"go.temporal.io/sdk/client"
)

type Dependency struct {
	Temporal   client.Client
	Card       card.CardWorkflow
	Repository *repository.Repositories
}

type Usecase struct {
	Cards
	SeattleWeather
}

func NewUsecase(deps Dependency) *Usecase {
	cardUsecase := NewCardUsecase(deps.Temporal, deps.Card)
	seattleWeatherUsecase := NewSeattleWeatherUsecase(deps.Repository)
	return &Usecase{
		Cards:          cardUsecase,
		SeattleWeather: seattleWeatherUsecase,
	}
}
