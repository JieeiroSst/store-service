package facade

import (
	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
)

type SeattleWeather interface {
	upsertBigQuerySeattleWeather(weathers []dto.SeattleWeatherRequestDTO)
	processSeattleWeather(weathers <-chan dto.SeattleWeatherRequestDTO, batchSize int)
	produceSeattleWeather(weathers []dto.SeattleWeatherRequestDTO, to chan dto.SeattleWeatherRequestDTO)
	InsertSeattleWeather(weathers []dto.SeattleWeatherRequestDTO)
}

type SeattleWeatherUsecase struct {
	repository *repository.Repositories
}

func NewSeattleWeatherUsecase(repository *repository.Repositories) *SeattleWeatherUsecase {
	return &SeattleWeatherUsecase{
		repository: repository,
	}
}

func (u *SeattleWeatherUsecase) upsertBigQuerySeattleWeather(weathers []dto.SeattleWeatherRequestDTO) {
}

func (u *SeattleWeatherUsecase) processSeattleWeather(weathers <-chan dto.SeattleWeatherRequestDTO, batchSize int) {
}

func (u *SeattleWeatherUsecase) produceSeattleWeather(weathers []dto.SeattleWeatherRequestDTO, to chan dto.SeattleWeatherRequestDTO) {
}

func (u *SeattleWeatherUsecase) InsertSeattleWeather(weathers []dto.SeattleWeatherRequestDTO) {}
