package usecase

import (
	"github.com/JIeeiroSst/workflow-service/internal/activities/approve"
	"github.com/JIeeiroSst/workflow-service/internal/activities/card"
	"go.temporal.io/sdk/client"
)

type Dependency struct {
	Temporal client.Client
	Card     card.CardWorkflow
	Approve  approve.ApproveWorkflow
}

type Usecase struct {
	Cards
	Approve
}

func NewUsecase(deps Dependency) *Usecase {
	cardUsecase := NewCardUsecase(deps.Temporal, deps.Card)
	approveUsecase := NewApproveUsecase(deps.Approve)
	return &Usecase{
		Cards:   cardUsecase,
		Approve: approveUsecase,
	}
}
