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
	ActiveUser
}

func NewFacde(deps Dependency) *Facade {
	seattleWeatherFacade := NewSeattleWeatherFacade(deps.Repository)
	gameFace := NewGameFacade(deps.Repository)
	spotifyQuarterly := NewSpotifyQuarterlyFacade(deps.Repository)
	sellingPlayStation := NewSellingPlayStationFacade(deps.Repository)
	return &Facade{
		SeattleWeather:     seattleWeatherFacade,
		Game:               gameFace,
		SpotifyQuarterly:   spotifyQuarterly,
		SellingPlayStation: sellingPlayStation,
	}
}

func (f *Facade) Factory(types TYPE) (resp interface{}) {
	switch types {
	case SeattleWeatherType:
		resp = f.SeattleWeather
	case GameType:
		resp = f.Game
	case SpotifyQuarterlyType:
		resp = f.SpotifyQuarterly
	case SellingPlayStationType:
		resp = f.SellingPlayStation
	default:
		fmt.Printf("%v\n", types)
	}
	return
}

// func convert(object interface{}) {
// 	user, ok := object.(SeattleWeather)

// 	if ok {
// 		fmt.Printf("Hello %s!\n", user)
// 	}
// }
