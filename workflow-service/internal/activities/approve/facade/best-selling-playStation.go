package facade

import (
	"fmt"
	"log"
	"sync"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"github.com/JIeeiroSst/workflow-service/model"
)

type SellingPlayStation interface {
	upsertBigQuerySellingPlayStation(playStations []dto.BestSellingPlayStationRequestDTO, batchID string)
	processSellingPlayStation(playStations <-chan dto.BestSellingPlayStationRequestDTO, batchSize int, batchID string)
	produceSellingPlayStation(playStations []dto.BestSellingPlayStationRequestDTO, to chan dto.BestSellingPlayStationRequestDTO, batchID string)
	InsertSellingPlayStation(playStations []dto.BestSellingPlayStationRequestDTO, batchID string)
}

type SellingPlayStationFacade struct {
	repository *repository.Repositories
}

func NewSellingPlayStationFacade(repository *repository.Repositories) *SellingPlayStationFacade {
	return &SellingPlayStationFacade{
		repository: repository,
	}
}

func (u *SellingPlayStationFacade) upsertBigQuerySellingPlayStation(playStations []dto.BestSellingPlayStationRequestDTO, batchID string) {
	fmt.Printf("Processing batch of %d\n", len(playStations))
	var playStationsModels []model.BestSellingPlayStation
	for _, playStation := range playStations {
		playStationsModel := u.mappingWeather(playStation)
		playStationsModels = append(playStationsModels, playStationsModel)
	}
	if err := u.repository.BestSellingPlayStations.InsertBatchBestSellingPlayStation(playStationsModels, batchID); err != nil {
		log.Println(err)
	}

}

func (u *SellingPlayStationFacade) processSellingPlayStation(weathers <-chan dto.BestSellingPlayStationRequestDTO, batchSize int, batchID string) {
	var batch []dto.BestSellingPlayStationRequestDTO
	for weather := range weathers {
		batch = append(batch, weather)
		if len(batch) == batchSize {
			u.upsertBigQuerySellingPlayStation(batch, batchID)
			batch = []dto.BestSellingPlayStationRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQuerySellingPlayStation(batch, batchID)
	}
}

func (u *SellingPlayStationFacade) produceSellingPlayStation(playStations []dto.BestSellingPlayStationRequestDTO, to chan dto.BestSellingPlayStationRequestDTO, batchID string) {
	for _, value := range playStations {
		to <- dto.BestSellingPlayStationRequestDTO{
			Game:        value.Game,
			CopiesSold:  value.CopiesSold,
			ReleaseDate: value.ReleaseDate,
			Genre:       value.Genre,
			Developer:   value.Developer,
			Publisher:   value.Publisher,
			BatchID:     batchID,
		}
	}
}

func (u *SellingPlayStationFacade) InsertSellingPlayStation(playStations []dto.BestSellingPlayStationRequestDTO, batchID string) {
	var wg sync.WaitGroup
	audits := make(chan dto.BestSellingPlayStationRequestDTO)
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.processSellingPlayStation(audits, batchSize, batchID)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.produceSellingPlayStation(playStations, audits, batchID)
		close(audits)
	}()
	wg.Wait()
	fmt.Println("Complete")
}

func (u *SellingPlayStationFacade) mappingWeather(playStation dto.BestSellingPlayStationRequestDTO) model.BestSellingPlayStation {
	return model.BestSellingPlayStation{
		Game:        playStation.Game,
		CopiesSold:  playStation.CopiesSold,
		ReleaseDate: playStation.ReleaseDate,
		Genre:       playStation.Genre,
		Developer:   playStation.Developer,
		Publisher:   playStation.Publisher,
	}
}
