package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type opportunityWorkflow struct {
	repo     port.Repository[model.Opportunity]
	notifier port.Notifier
}

func NewOpportunityWorkflow(repo port.Repository[model.Opportunity], n port.Notifier) port.OpportunityWorkflow {
	return &opportunityWorkflow{repo: repo, notifier: n}
}

func (w *opportunityWorkflow) ChangeStage(ctx context.Context, id uint, stage string) (*model.Opportunity, error) {
	if !model.ValidStage(stage) {
		return nil, common.ErrInvalidRequest
	}
	opp, err := w.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if opp.Stage == model.StageWon || opp.Stage == model.StageLost {
		return nil, common.ErrInvalidTransition
	}

	opp.Stage = stage
	if err := w.repo.Update(ctx, id, opp); err != nil {
		return nil, err
	}
	if stage == model.StageWon {
		notify(ctx, w.notifier, "Opportunity won", "Opportunity #%d (%.2f) was won", opp.ID, opp.Amount)
	}
	return w.repo.GetByID(ctx, id)
}

func (w *contractWorkflow) Submit(ctx context.Context, id uint) (*model.Contract, error) {
	return w.move(ctx, id, model.ContractDraft, model.ContractPending, "")
}

func (w *contractWorkflow) Approve(ctx context.Context, id uint) (*model.Contract, error) {
	c, err := w.move(ctx, id, model.ContractPending, model.ContractApproved, model.ContractApproved)
	if err == nil {
		notify(ctx, w.notifier, "Contract approved", "Contract #%d for account #%d was approved", c.ID, c.AccountID)
	}
	return c, err
}

func (w *contractWorkflow) Reject(ctx context.Context, id uint) (*model.Contract, error) {
	return w.move(ctx, id, model.ContractPending, model.ContractRejected, model.ContractRejected)
}

func (w *contractWorkflow) SigningPayload(ctx context.Context, id uint, signedBy string, endDate, signingTime time.Time) ([]byte, error) {
	c, err := w.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return model.SigningPayload(c, signedBy, endDate.UTC().Truncate(time.Second), signingTime.UTC().Truncate(time.Second)), nil
}

func (w *contractWorkflow) move(ctx context.Context, id uint, from, to, approval string) (*model.Contract, error) {
	c, err := w.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != from {
		return nil, common.ErrInvalidTransition
	}

	if to != model.ContractRejected && c.EndDate != nil && !c.EndDate.After(w.now()) {
		return nil, common.ErrInvalidTransition
	}
	c.Status, c.Approval = to, approval
	if err := w.repo.Update(ctx, id, c); err != nil {
		return nil, err
	}
	syncLifecycle(ctx, w.orchestrator, id)
	return w.repo.GetByID(ctx, id)
}

type caseWorkflow struct {
	repo     port.Repository[model.Case]
	notifier port.Notifier
}

func NewCaseWorkflow(repo port.Repository[model.Case], n port.Notifier) port.CaseWorkflow {
	return &caseWorkflow{repo: repo, notifier: n}
}

func (w *caseWorkflow) Close(ctx context.Context, id uint) (*model.Case, error) {
	c, err := w.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status == model.CaseClosed {
		return nil, common.ErrInvalidTransition
	}

	now := time.Now()
	c.Status, c.ClosedAt = model.CaseClosed, &now
	if err := w.repo.Update(ctx, id, c); err != nil {
		return nil, err
	}
	notify(ctx, w.notifier, "Case closed", "Case #%d was closed", c.ID)
	return w.repo.GetByID(ctx, id)
}
