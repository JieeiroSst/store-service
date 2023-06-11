package facade

import (
	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
)

type SpotifyQuarterly interface {
	upsertBigQuerySpotifyQuarterly(weathers []dto.SpotifyQuarterlyRequestDTO)
	processSpotifyQuarterly(weathers <-chan dto.SpotifyQuarterlyRequestDTO, batchSize int)
	produceSpotifyQuarterly(weathers []dto.SpotifyQuarterlyRequestDTO, to chan dto.SpotifyQuarterlyRequestDTO)
	InsertSpotifyQuarterly(weathers []dto.SpotifyQuarterlyRequestDTO)
}

type SpotifyQuarterlyFacade struct {
	repository *repository.Repositories
}

func NewSpotifyQuarterlyFacade(repository *repository.Repositories) *SellingPlayStationFacade {
	return &SellingPlayStationFacade{
		repository: repository,
	}
}

func (u *SellingPlayStationFacade) upsertBigQuerySpotifyQuarterly(weathers []dto.SpotifyQuarterlyRequestDTO) {

}

func (u *SellingPlayStationFacade) processSpotifyQuarterly(weathers <-chan dto.SpotifyQuarterlyRequestDTO, batchSize int) {
}

func (u *SellingPlayStationFacade) produceSpotifyQuarterly(weathers []dto.SpotifyQuarterlyRequestDTO, to chan dto.SpotifyQuarterlyRequestDTO) {
}

func (u *SellingPlayStationFacade) InsertSpotifyQuarterly(weathers []dto.SpotifyQuarterlyRequestDTO) {
}
