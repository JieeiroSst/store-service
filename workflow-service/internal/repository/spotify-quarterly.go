package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/workflow-service/model"
)

type SpotifyQuarterlys interface {
	InsertBatchSpotifyQuarterlys(spotifyQuarterlys []model.SpotifyQuarterly) error
}

type SpotifyQuarterlyRepo struct {
	DB *sql.DB
}

func NewSpotifyQuarterlyRepo(db *sql.DB) *SpotifyQuarterlyRepo {
	return &SpotifyQuarterlyRepo{
		DB: db,
	}
}

func (r *SpotifyQuarterlyRepo) InsertBatchSpotifyQuarterlys(spotifyQuarterlys []model.SpotifyQuarterly) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, spotifyQuarterly := range spotifyQuarterlys {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, spotifyQuarterly.Date)
		valueArgs = append(valueArgs, spotifyQuarterly.TotalRevenue)
		valueArgs = append(valueArgs, spotifyQuarterly.CostOfRevenue)
		valueArgs = append(valueArgs, spotifyQuarterly.GrossProfit)
		valueArgs = append(valueArgs, spotifyQuarterly.PremiumRevenue)
		valueArgs = append(valueArgs, spotifyQuarterly.PremiumCostRevenue)
		valueArgs = append(valueArgs, spotifyQuarterly.PremiumGrossProfit)
		valueArgs = append(valueArgs, spotifyQuarterly.AdRevenue)
		valueArgs = append(valueArgs, spotifyQuarterly.AdCostOfRevenue)
		valueArgs = append(valueArgs, spotifyQuarterly.AdGrossProfit)
		valueArgs = append(valueArgs, spotifyQuarterly.MAUs)
		valueArgs = append(valueArgs, spotifyQuarterly.PremiumARPU)
		valueArgs = append(valueArgs, spotifyQuarterly.AdMAUs)
		valueArgs = append(valueArgs, spotifyQuarterly.PremiumARPU)
		valueArgs = append(valueArgs, spotifyQuarterly.SalesAndMarketingCost)
		valueArgs = append(valueArgs, spotifyQuarterly.ResearchAndDevelopmentCost)
		valueArgs = append(valueArgs, spotifyQuarterly.GenrealAndAdminstraiveCost)
	}

	smt := `INSERT INTO spotify_quarterlies(date,total_revenue,cost_of_revenue,gross_profit,
		premium_revenue,premium_cost_revenue,premium_gross_profit,ad_revenue,
		ad_cost_of_revenue,ad_gross_profit,maus,premium_maus,ad_maus,	
		premium_arpu,sales_and_marketing_cost,research_and_development_cost,
		genreal_and_adminstraive_cost) VALUES %s`

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
