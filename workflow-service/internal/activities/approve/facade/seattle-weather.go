package facade

import (
	"fmt"
	"log"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"github.com/JIeeiroSst/workflow-service/model"
)

type SeattleWeather interface {
	upsertBigQuerySeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, batchID string)
	processSeattleWeather(seattleWeathers <-chan dto.SeattleWeatherRequestDTO, batchSize int, batchID string)
	produceSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, to chan dto.SeattleWeatherRequestDTO, batchID string)
	InsertSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, batchID string)
}

type SeattleWeatherFacade struct {
	repository *repository.Repositories
}

func NewSeattleWeatherFacade(repository *repository.Repositories) *SeattleWeatherFacade {
	return &SeattleWeatherFacade{
		repository: repository,
	}
}

func (u *SeattleWeatherFacade) upsertBigQuerySeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, batchID string) {
	fmt.Printf("Processing batch of %d\n", len(seattleWeathers))
	var seattleWeathersModels []model.SeattleWeather
	for _, game := range seattleWeathers {
		seattleWeathersModel := u.mappingSeattleWeather(game)
		seattleWeathersModels = append(seattleWeathersModels, seattleWeathersModel)
	}
	if err := u.repository.SeattleWeathers.InsertBatchSeattleWeathers(seattleWeathersModels, batchID); err != nil {
		log.Println(err)
	}
}

func (u *SeattleWeatherFacade) processSeattleWeather(seattleWeathers <-chan dto.SeattleWeatherRequestDTO, batchSize int, batchID string) {
	var batch []dto.SeattleWeatherRequestDTO
	for seattleWeather := range seattleWeathers {
		batch = append(batch, seattleWeather)
		if len(batch) == batchSize {
			u.upsertBigQuerySeattleWeather(batch, batchID)
			batch = []dto.SeattleWeatherRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQuerySeattleWeather(batch, batchID)
	}
}

func (u *SeattleWeatherFacade) produceSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, to chan dto.SeattleWeatherRequestDTO, batchID string) {
	for _, seattleWeather := range seattleWeathers {
		to <- dto.SeattleWeatherRequestDTO{
			Date:          seattleWeather.Date,
			Precipitation: seattleWeather.Precipitation,
			TempMax:       seattleWeather.TempMax,
			TempMin:       seattleWeather.TempMin,
			Wind:          seattleWeather.Wind,
			Weather:       seattleWeather.Weather,
			BatchID:       batchID,
		}
	}
}

func (u *SeattleWeatherFacade) InsertSeattleWeather(seattleWeathers []dto.SeattleWeatherRequestDTO, batchID string) {
	var batch []dto.SeattleWeatherRequestDTO
	for _, seattleWeather := range seattleWeathers {
		batch = append(batch, seattleWeather)
		if len(batch) == batchSize {
			u.upsertBigQuerySeattleWeather(batch, batchID)
			batch = []dto.SeattleWeatherRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQuerySeattleWeather(batch, batchID)
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
