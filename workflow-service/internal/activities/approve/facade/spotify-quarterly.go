package facade

import (
	"fmt"
	"log"
	"sync"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"github.com/JIeeiroSst/workflow-service/model"
)

type SpotifyQuarterly interface {
	upsertBigQuerySpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, batchID string)
	processSpotifyQuarterly(spotifyQuarterlies <-chan dto.SpotifyQuarterlyRequestDTO, batchSize int, batchID string)
	produceSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, to chan dto.SpotifyQuarterlyRequestDTO, batchID string)
	InsertSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, batchID string)
}

type SpotifyQuarterlyFacade struct {
	repository *repository.Repositories
}

func NewSpotifyQuarterlyFacade(repository *repository.Repositories) *SellingPlayStationFacade {
	return &SellingPlayStationFacade{
		repository: repository,
	}
}

func (u *SellingPlayStationFacade) upsertBigQuerySpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, batchID string) {
	fmt.Printf("Processing batch of %d\n", len(spotifyQuarterlies))
	var spotifyQuarterlyModels []model.SpotifyQuarterly
	for _, spotifyQuarterly := range spotifyQuarterlies {
		spotifyQuarterlyModel := u.mappingSpotifyQuarterly(spotifyQuarterly)
		spotifyQuarterlyModels = append(spotifyQuarterlyModels, spotifyQuarterlyModel)
	}
	if err := u.repository.SpotifyQuarterlys.InsertBatchSpotifyQuarterlys(spotifyQuarterlyModels, batchID); err != nil {
		log.Println(err)
	}
}

func (u *SellingPlayStationFacade) processSpotifyQuarterly(spotifyQuarterlies <-chan dto.SpotifyQuarterlyRequestDTO, batchSize int, batchID string) {
	var batch []dto.SpotifyQuarterlyRequestDTO
	for spotifyQuarterlie := range spotifyQuarterlies {
		batch = append(batch, spotifyQuarterlie)
		if len(batch) == batchSize {
			u.upsertBigQuerySpotifyQuarterly(batch, batchID)
			batch = []dto.SpotifyQuarterlyRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQuerySpotifyQuarterly(batch, batchID)
	}
}

func (u *SellingPlayStationFacade) produceSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, to chan dto.SpotifyQuarterlyRequestDTO, batchID string) {
	for _, spotifyQuarterly := range spotifyQuarterlies {
		to <- dto.SpotifyQuarterlyRequestDTO{
			Date:                       spotifyQuarterly.Date,
			TotalRevenue:               spotifyQuarterly.TotalRevenue,
			CostOfRevenue:              spotifyQuarterly.CostOfRevenue,
			GrossProfit:                spotifyQuarterly.GrossProfit,
			PremiumRevenue:             spotifyQuarterly.PremiumRevenue,
			PremiumCostRevenue:         spotifyQuarterly.PremiumCostRevenue,
			PremiumGrossProfit:         spotifyQuarterly.PremiumGrossProfit,
			AdRevenue:                  spotifyQuarterly.AdRevenue,
			AdCostOfRevenue:            spotifyQuarterly.AdCostOfRevenue,
			AdGrossProfit:              spotifyQuarterly.AdGrossProfit,
			MAUs:                       spotifyQuarterly.MAUs,
			PremiumARPU:                spotifyQuarterly.PremiumARPU,
			AdMAUs:                     spotifyQuarterly.AdMAUs,
			SalesAndMarketingCost:      spotifyQuarterly.SalesAndMarketingCost,
			ResearchAndDevelopmentCost: spotifyQuarterly.ResearchAndDevelopmentCost,
			GenrealAndAdminstraiveCost: spotifyQuarterly.GenrealAndAdminstraiveCost,
			BatchID:                    batchID,
		}
	}
}

func (u *SellingPlayStationFacade) InsertSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, batchID string) {
	var wg sync.WaitGroup
	audits := make(chan dto.SpotifyQuarterlyRequestDTO)
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.processSpotifyQuarterly(audits, batchSize, batchID)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.produceSpotifyQuarterly(spotifyQuarterlies, audits, batchID)
		close(audits)
	}()
	wg.Wait()
	fmt.Println("Complete")
}

func (u *SellingPlayStationFacade) mappingSpotifyQuarterly(spotifyQuarterly dto.SpotifyQuarterlyRequestDTO) model.SpotifyQuarterly {
	return model.SpotifyQuarterly{
		Date:                       spotifyQuarterly.Date,
		TotalRevenue:               spotifyQuarterly.TotalRevenue,
		CostOfRevenue:              spotifyQuarterly.CostOfRevenue,
		GrossProfit:                spotifyQuarterly.GrossProfit,
		PremiumRevenue:             spotifyQuarterly.PremiumRevenue,
		PremiumCostRevenue:         spotifyQuarterly.PremiumCostRevenue,
		PremiumGrossProfit:         spotifyQuarterly.PremiumGrossProfit,
		AdRevenue:                  spotifyQuarterly.AdRevenue,
		AdCostOfRevenue:            spotifyQuarterly.AdCostOfRevenue,
		AdGrossProfit:              spotifyQuarterly.AdGrossProfit,
		MAUs:                       spotifyQuarterly.MAUs,
		PremiumARPU:                spotifyQuarterly.PremiumARPU,
		AdMAUs:                     spotifyQuarterly.AdMAUs,
		SalesAndMarketingCost:      spotifyQuarterly.SalesAndMarketingCost,
		ResearchAndDevelopmentCost: spotifyQuarterly.ResearchAndDevelopmentCost,
		GenrealAndAdminstraiveCost: spotifyQuarterly.GenrealAndAdminstraiveCost,
	}
}
