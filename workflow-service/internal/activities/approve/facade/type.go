package facade

import "fmt"

type TYPE string

const (
	SeattleWeatherType     TYPE = "seattle_weather_type"
	GameType               TYPE = "game_type"
	SpotifyQuarterlyType   TYPE = "spotify_quarterly_type"
	SellingPlayStationType TYPE = "selling_playstation_type"
)

func (t TYPE) String() string {
	switch t {
	case SeattleWeatherType:
		return "seattle_weather_type"
	case GameType:
		return "game_type"
	case SpotifyQuarterlyType:
		return "spotify_quarterly_type"
	case SellingPlayStationType:
		return "selling_playstation_type"
	default:
		return fmt.Sprintf("%d", string(t))
	}
}
