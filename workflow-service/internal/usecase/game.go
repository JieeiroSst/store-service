package usecase

import (
	"fmt"
	"log"
	"sync"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
	"github.com/JIeeiroSst/workflow-service/model"
)

type Game interface {
	upsertBigQueryGame(weathers []dto.GameRequestDTO)
	processGame(weathers <-chan dto.GameRequestDTO, batchSize int)
	produceGame(weathers []dto.GameRequestDTO, to chan dto.GameRequestDTO)
	InsertGame(weathers []dto.GameRequestDTO)
}

type GameUsecase struct {
	repository repository.Repositories
}

func NewGameUsecase(repository repository.Repositories) *GameUsecase {
	return &GameUsecase{
		repository: repository,
	}
}

func (u *GameUsecase) upsertBigQueryGame(games []dto.GameRequestDTO) {
	fmt.Printf("Processing batch of %d\n", len(games))
	var gameModels []model.Game
	for _, game := range games {
		gameModel := u.mappingGame(game)
		gameModels = append(gameModels, gameModel)
	}
	if err := u.repository.Games.InsertBatchGame(gameModels); err != nil {
		log.Println(err)
	}
}

func (u *GameUsecase) processGame(weathers <-chan dto.GameRequestDTO, batchSize int) {
	var batch []dto.GameRequestDTO
	for weather := range weathers {
		batch = append(batch, weather)
		if len(batch) == batchSize {
			u.upsertBigQueryGame(batch)
			batch = []dto.GameRequestDTO{}
		}
	}
	if len(batch) > 0 {
		u.upsertBigQueryGame(batch)
	}
}

func (u *GameUsecase) produceGame(games []dto.GameRequestDTO, to chan dto.GameRequestDTO) {
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
		}
	}
}

const batchSize = 1000

func (u *GameUsecase) InsertGame(weathers []dto.GameRequestDTO) {
	var wg sync.WaitGroup
	audits := make(chan dto.GameRequestDTO)
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.processGame(audits, batchSize)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		u.produceGame(weathers, audits)
		close(audits)
	}()
	wg.Wait()
	fmt.Println("Complete")
}

func (u *GameUsecase) mappingGame(game dto.GameRequestDTO) model.Game {
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
