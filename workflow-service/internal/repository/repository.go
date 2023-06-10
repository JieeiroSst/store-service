package repository

import "database/sql"

type Repositories struct {
	SpotifyQuarterlys
	Games
	BestSellingPlayStations
	SeattleWeathers
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		SpotifyQuarterlys:       NewSpotifyQuarterlyRepo(db),
		Games:                   NewGameRepo(db),
		BestSellingPlayStations: NewBestSellingPlayStationRepo(db),
		SeattleWeathers:         NewSeattleWeatherRepo(db),
	}
}
