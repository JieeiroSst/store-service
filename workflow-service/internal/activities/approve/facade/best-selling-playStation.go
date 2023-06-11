package facade

import (
	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
)

type SellingPlayStation interface {
	upsertBigQuerySellingPlayStation(weathers []dto.BestSellingPlayStationRequestDTO)
	processSellingPlayStation(weathers <-chan dto.BestSellingPlayStationRequestDTO, batchSize int)
	produceSellingPlayStation(weathers []dto.BestSellingPlayStationRequestDTO, to chan dto.BestSellingPlayStationRequestDTO)
	InsertSellingPlayStation(weathers []dto.BestSellingPlayStationRequestDTO)
}

type SellingPlayStationFacade struct {
	repository *repository.Repositories
}

func NewSellingPlayStationFacade(repository *repository.Repositories) *SellingPlayStationFacade {
	return &SellingPlayStationFacade{
		repository: repository,
	}
}

func (u *SellingPlayStationFacade) upsertBigQuerySellingPlayStation(weathers []dto.BestSellingPlayStationRequestDTO) {

}

func (u *SellingPlayStationFacade) processSellingPlayStation(weathers <-chan dto.BestSellingPlayStationRequestDTO, batchSize int) {
}

func (u *SellingPlayStationFacade) produceSellingPlayStation(weathers []dto.BestSellingPlayStationRequestDTO, to chan dto.BestSellingPlayStationRequestDTO) {
}

func (u *SellingPlayStationFacade) InsertSellingPlayStation(weathers []dto.BestSellingPlayStationRequestDTO) {
}
