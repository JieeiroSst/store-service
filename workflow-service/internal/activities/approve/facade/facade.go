package facade

import (
	"fmt"

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
	SpotifyQuarterly
	SellingPlayStation
}

func NewFacde(deps Dependency, types TYPE) (resp interface{}) {
	switch types {
	case SeattleWeatherType:
		seattleWeatherFacade := NewSeattleWeatherFacade(deps.Repository)
		resp = seattleWeatherFacade
	case GameType:
		gameFace := NewGameFacade(deps.Repository)
		resp = gameFace
	case SpotifyQuarterlyType:
		spotifyQuarterly := NewSpotifyQuarterlyFacade(deps.Repository)
		resp = spotifyQuarterly
	case SellingPlayStationType:
		sellingPlayStation := NewSellingPlayStationFacade(deps.Repository)
		resp = sellingPlayStation
	default:
		fmt.Printf("%v\n", types)
	}
	return
}
