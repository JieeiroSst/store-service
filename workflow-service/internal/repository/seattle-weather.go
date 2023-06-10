package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/workflow-service/model"
)

type SeattleWeathers interface {
	InsertBatchSeattleWeathers(seattleWeathers []model.SeattleWeather) error
}

type SeattleWeatherRepo struct {
	DB *sql.DB
}

func NewSeattleWeatherRepo(db *sql.DB) *SeattleWeatherRepo {
	return &SeattleWeatherRepo{
		DB: db,
	}
}

func (r *SeattleWeatherRepo) InsertBatchSeattleWeathers(seattleWeathers []model.SeattleWeather) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, seattleWeather := range seattleWeathers {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, seattleWeather.Date)
		valueArgs = append(valueArgs, seattleWeather.Precipitation)
		valueArgs = append(valueArgs, seattleWeather.TempMax)
		valueArgs = append(valueArgs, seattleWeather.TempMin)
		valueArgs = append(valueArgs, seattleWeather.Wind)
		valueArgs = append(valueArgs, seattleWeather.Weather)
	}

	smt := `INSERT INTO seattle_weather(date,precipitation,temp_max,temp_min,wind,weather) VALUES %s`

	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))
	tx, _ := r.DB.Begin()
	row, err := tx.Exec(smt, valueArgs...)
	if err != nil {
		tx.Rollback()
		return err
	}
	fmt.Println(row.LastInsertId())
	return nil
}
