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
	upsertBigQuerySpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO)
	processSpotifyQuarterly(spotifyQuarterlies <-chan dto.SpotifyQuarterlyRequestDTO, batchSize int)
	produceSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, to chan dto.SpotifyQuarterlyRequestDTO)
	InsertSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO)
}

type SpotifyQuarterlyFacade struct {
	repository *repository.Repositories
}

func NewSpotifyQuarterlyFacade(repository *repository.Repositories) *SellingPlayStationFacade {
	return &SellingPlayStationFacade{
		repository: repository,
	}
}

func (u *SellingPlayStationFacade) upsertBigQuerySpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO) {
	fmt.Printf("Processing batch of %d\n", len(spotifyQuarterlies))
	var spotifyQuarterlyModels []model.SpotifyQuarterly
	for _, spotifyQuarterly := range spotifyQuarterlies {
		spotifyQuarterlyModel := u.mappingSpotifyQuarterly(spotifyQuarterly)
		spotifyQuarterlyModels = append(spotifyQuarterlyModels, spotifyQuarterlyModel)
	}
	if err := u.repository.SpotifyQuarterlys.InsertBatchSpotifyQuarterlys(spotifyQuarterlyModels); err != nil {
		log.Println(err)
	}
}

func (u *SellingPlayStationFacade) processSpotifyQuarterly(spotifyQuarterlies <-chan dto.SpotifyQuarterlyRequestDTO, batchSize int) {
	var batch []dto.SpotifyQuarterlyRequestDTO
	for spotifyQuarterlie := range spotifyQuarterlies {
		batch = append(batch, spotifyQuarterlie)
		if len(batch) == batchSize {
			u.upsertBigQuerySpotifyQuarterly(batch)
			batch = []dto.SpotifyQuarterlyRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQuerySpotifyQuarterly(batch)
	}
}

func (u *SellingPlayStationFacade) produceSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO, to chan dto.SpotifyQuarterlyRequestDTO) {
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
		}
	}
}

func (u *SellingPlayStationFacade) InsertSpotifyQuarterly(spotifyQuarterlies []dto.SpotifyQuarterlyRequestDTO) {
	var wg sync.WaitGroup
	audits := make(chan dto.SpotifyQuarterlyRequestDTO)
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.processSpotifyQuarterly(audits, batchSize)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.produceSpotifyQuarterly(spotifyQuarterlies, audits)
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
