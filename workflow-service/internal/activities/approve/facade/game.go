package facade

import (
	"fmt"
	"log"
	"sync"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"github.com/JIeeiroSst/workflow-service/model"
)

type Game interface {
	upsertBigQueryGame(weathers []dto.GameRequestDTO, batchID string)
	processGame(weathers <-chan dto.GameRequestDTO, batchSize int, batchID string)
	produceGame(weathers []dto.GameRequestDTO, to chan dto.GameRequestDTO, batchID string)
	InsertGame(weathers []dto.GameRequestDTO, batchID string)
}

type GameFace struct {
	repository *repository.Repositories
}

func NewGameFacade(repository *repository.Repositories) *GameFace {
	return &GameFace{
		repository: repository,
	}
}

func (u *GameFace) upsertBigQueryGame(games []dto.GameRequestDTO, batchID string) {
	fmt.Printf("Processing batch of %d\n", len(games))
	var gameModels []model.Game
	for _, game := range games {
		gameModel := u.mappingGame(game)
		gameModels = append(gameModels, gameModel)
	}
	if err := u.repository.Games.InsertBatchGame(gameModels, batchID); err != nil {
		log.Println(err)
	}
}

func (u *GameFace) processGame(weathers <-chan dto.GameRequestDTO, batchSize int, batchID string) {
	var batch []dto.GameRequestDTO
	for weather := range weathers {
		batch = append(batch, weather)
		if len(batch) == batchSize {
			u.upsertBigQueryGame(batch, batchID)
			batch = []dto.GameRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQueryGame(batch, batchID)
	}
}

func (u *GameFace) produceGame(games []dto.GameRequestDTO, to chan dto.GameRequestDTO, batchID string) {
	for _, game := range games {
		to <- dto.GameRequestDTO{
			ID:            game.ID,
			Rated:         game.Rated,
			CreatedAt:     game.CreatedAt,
			LastMoveAt:    game.LastMoveAt,
			Turns:         game.Turns,
			VictoryStatus: game.VictoryStatus,
			Winner:        game.Winner,
			IncrementCode: game.IncrementCode,
			WhiteId:       game.WhiteId,
			WhiteRating:   game.WhiteRating,
			BlackId:       game.BlackId,
			BlackRating:   game.BlackRating,
			Moves:         game.Moves,
			OpeningEco:    game.OpeningEco,
			OpeningName:   game.OpeningName,
			OpeningPly:    game.OpeningPly,
			BatchID:       batchID,
		}
	}
}

const batchSize = 1000

func (u *GameFace) InsertGame(weathers []dto.GameRequestDTO, batchID string) {
	var wg sync.WaitGroup
	audits := make(chan dto.GameRequestDTO)
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.processGame(audits, batchSize, batchID)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.produceGame(weathers, audits, batchID)
		close(audits)
	}()
	wg.Wait()
	fmt.Println("Complete")
}

func (u *GameFace) mappingGame(game dto.GameRequestDTO) model.Game {
	return model.Game{
		ID:            game.ID,
		Rated:         game.Rated,
		CreatedAt:     game.CreatedAt,
		LastMoveAt:    game.LastMoveAt,
		Turns:         game.Turns,
		VictoryStatus: game.VictoryStatus,
		Winner:        game.Winner,
		IncrementCode: game.IncrementCode,
		WhiteId:       game.WhiteId,
		WhiteRating:   game.WhiteRating,
		BlackId:       game.BlackId,
		BlackRating:   game.BlackRating,
		Moves:         game.Moves,
		OpeningEco:    game.OpeningEco,
		OpeningName:   game.OpeningName,
		OpeningPly:    game.OpeningPly,
	}
}
