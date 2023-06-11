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

type SeattleWeatherFacade struct {
	repository *repository.Repositories
}

func NewSeattleWeatherFacade(repository *repository.Repositories) *SeattleWeatherFacade {
	return &SeattleWeatherFacade{
		repository: repository,
	}
}

func (u *SeattleWeatherFacade) upsertBigQuerySeattleWeather(weathers []dto.SeattleWeatherRequestDTO) {

}

func (u *SeattleWeatherFacade) processSeattleWeather(weathers <-chan dto.SeattleWeatherRequestDTO, batchSize int) {
}

func (u *SeattleWeatherFacade) produceSeattleWeather(weathers []dto.SeattleWeatherRequestDTO, to chan dto.SeattleWeatherRequestDTO) {
}

func (u *SeattleWeatherFacade) InsertSeattleWeather(weathers []dto.SeattleWeatherRequestDTO) {}
