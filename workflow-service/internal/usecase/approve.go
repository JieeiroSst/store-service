package usecase

import (
	"github.com/JIeeiroSst/workflow-service/internal/activities/approve"
)

type Approve interface {
}

type ApproveUsecase struct {
	approve approve.ApproveWorkflow
}

func NewApproveUsecase(approve approve.ApproveWorkflow) *ApproveUsecase {
	return &ApproveUsecase{
		approve: approve,
	}
}
