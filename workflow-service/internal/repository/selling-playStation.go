package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/workflow-service/model"
)

type BestSellingPlayStations interface {
	InsertBatchBestSellingPlayStation(sellingPlayStations []model.BestSellingPlayStation) error
}

type BestSellingPlayStationRepo struct {
	DB *sql.DB
}

func NewBestSellingPlayStationRepo(db *sql.DB) *BestSellingPlayStationRepo {
	return &BestSellingPlayStationRepo{
		DB: db,
	}
}

func (r *BestSellingPlayStationRepo) InsertBatchBestSellingPlayStation(sellingPlayStations []model.BestSellingPlayStation) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, sellingPlayStation := range sellingPlayStations {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, sellingPlayStation.Game)
		valueArgs = append(valueArgs, sellingPlayStation.CopiesSold)
		valueArgs = append(valueArgs, sellingPlayStation.ReleaseDate)
		valueArgs = append(valueArgs, sellingPlayStation.Genre)
		valueArgs = append(valueArgs, sellingPlayStation.Developer)
		valueArgs = append(valueArgs, sellingPlayStation.Publisher)
	}

	smt := `INSERT INTO best_selling_playStations(game,copies_sold,release_date,genre,developer,publisher) VALUES %s`

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
