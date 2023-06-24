package facade

import (
	"fmt"
	"log"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"github.com/JIeeiroSst/workflow-service/model"
)

type SeattleWeather interface {
	upsertBigQuerySeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO)
	processSeattleWeather(seattleWeathers <-chan dto.SeattleWeatherRequestDTO, batchSize int)
	produceSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, to chan dto.SeattleWeatherRequestDTO)
	InsertSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO)
}

type SeattleWeatherFacade struct {
	repository *repository.Repositories
}

func NewSeattleWeatherFacade(repository *repository.Repositories) *SeattleWeatherFacade {
	return &SeattleWeatherFacade{
		repository: repository,
	}
}

func (u *SeattleWeatherFacade) upsertBigQuerySeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO) {
	fmt.Printf("Processing batch of %d\n", len(seattleWeathers))
	var seattleWeathersModels []model.SeattleWeather
	for _, game := range seattleWeathers {
		seattleWeathersModel := u.mappingSeattleWeather(game)
		seattleWeathersModels = append(seattleWeathersModels, seattleWeathersModel)
	}
	if err := u.repository.SeattleWeathers.InsertBatchSeattleWeathers(seattleWeathersModels); err != nil {
		log.Println(err)
	}
}

func (u *SeattleWeatherFacade) processSeattleWeather(seattleWeathers <-chan dto.SeattleWeatherRequestDTO, batchSize int) {
	var batch []dto.SeattleWeatherRequestDTO
	for seattleWeather := range seattleWeathers {
		batch = append(batch, seattleWeather)
		if len(batch) == batchSize {
			u.upsertBigQuerySeattleWeather(batch)
			batch = []dto.SeattleWeatherRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQuerySeattleWeather(batch)
	}
}

func (u *SeattleWeatherFacade) produceSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, to chan dto.SeattleWeatherRequestDTO) {
	for _, seattleWeather := range seattleWeathers {
		to <- dto.SeattleWeatherRequestDTO{
			Date:          seattleWeather.Date,
			Precipitation: seattleWeather.Precipitation,
			TempMax:       seattleWeather.TempMax,
			TempMin:       seattleWeather.TempMin,
			Wind:          seattleWeather.Wind,
			Weather:       seattleWeather.Weather,
		}
	}
}

func (u *SeattleWeatherFacade) InsertSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO) {
	var batch []dto.SeattleWeatherRequestDTO
	for _, seattleWeather := range seattleWeathers {
		batch = append(batch, seattleWeather)
		if len(batch) == batchSize {
			u.upsertBigQuerySeattleWeather(batch)
			batch = []dto.SeattleWeatherRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQuerySeattleWeather(batch)
	}
}

func (u *SeattleWeatherFacade) mappingSeattleWeather(seattleWeather dto.SeattleWeatherRequestDTO) model.SeattleWeather {
	return model.SeattleWeather{
		Date:          seattleWeather.Date,
		Precipitation: seattleWeather.Precipitation,
		TempMax:       seattleWeather.TempMax,
		TempMin:       seattleWeather.TempMin,
		Wind:          seattleWeather.Wind,
		Weather:       seattleWeather.Weather,
	}
}
