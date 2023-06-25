package approve

import (
	"context"

	"github.com/JIeeiroSst/workflow-service/internal/activities/approve/facade"
)

type ProcessState struct {
	Facade facade.Facade
}

type Upload struct {
	Type         string
	File         string
	ProcessTable ProcessState
}

type Process struct {
	Type         string
	ProcessState ProcessState
	Email        string
	IsApprove    bool
}

type Approve struct {
	Type         string
	ProcessState ProcessState
	Email        string
	IsApprove    bool
}

func (p *ProcessState) UploadApprove(upload Upload) {

}

func (p *ProcessState) ProcessApprove(process Process) {

}

func (a *ProcessState) ApproveProcess(_ context.Context, process ProcessState) error {

	return nil
}

func (a *ProcessState) SendAbandonedProcess(_ context.Context, isApprove bool) error {

	return nil
}
