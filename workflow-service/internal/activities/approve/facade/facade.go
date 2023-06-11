package facade

import (
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"go.temporal.io/sdk/client"
)

type Dependency struct {
	Temporal   client.Client
	Repository *repository.Repositories
}

type Facade struct {
	SeattleWeather
	Game
}

func NewFacde(deps Dependency) *Facade {
	seattleWeatherFacade := NewSeattleWeatherFacade(deps.Repository)
	gameFace := NewGameFacade(deps.Repository)
	return &Facade{
		SeattleWeather: seattleWeatherFacade,
		Game:           gameFace,
	}
}
