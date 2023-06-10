package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/workflow-service/model"
)

type Games interface {
	InsertBatchGame(games []model.Game) error
}

type GameRepo struct {
	DB *sql.DB
}

func NewGameRepo(db *sql.DB) *GameRepo {
	return &GameRepo{
		DB: db,
	}
}

func (r *GameRepo) InsertBatchGame(games []model.Game) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, game := range games {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, game.ID)
		valueArgs = append(valueArgs, game.Rated)
		valueArgs = append(valueArgs, game.CreatedAt)
		valueArgs = append(valueArgs, game.LastMoveAt)
		valueArgs = append(valueArgs, game.Turns)
		valueArgs = append(valueArgs, game.VictoryStatus)
		valueArgs = append(valueArgs, game.Winner)
		valueArgs = append(valueArgs, game.IncrementCode)
		valueArgs = append(valueArgs, game.WhiteId)
		valueArgs = append(valueArgs, game.WhiteRating)
		valueArgs = append(valueArgs, game.BlackId)
		valueArgs = append(valueArgs, game.BlackRating)
		valueArgs = append(valueArgs, game.Moves)
		valueArgs = append(valueArgs, game.OpeningEco)
		valueArgs = append(valueArgs, game.OpeningName)
		valueArgs = append(valueArgs, game.OpeningPly)
	}

	smt := `INSERT INTO games(id,rated,created_at,last_move_at,turns,victory_status,winner,
		increment_code,white_id,white_rating,black_i,black_rating,moves,opening_eco,
		opening_name,opening_ply) VALUES %s`

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
