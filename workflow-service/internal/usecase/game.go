package usecase

import (
	"github.com/JIeeiroSst/workflow-service/internal/activities/approve"
)

type Game interface {
}

type GameUsecase struct {
	approve approve.ApproveWorkflow
}

func NewGameFacade(approve approve.ApproveWorkflow) *GameUsecase {
	return &GameUsecase{
		approve: approve,
	}
}
