package application

import (
	"context"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

func NewLeadService(repo port.Repository[model.Lead], n port.Notifier) port.Usecase[model.Lead] {
	return newCRUD(repo, hooks[model.Lead]{
		prepare: func(l *model.Lead) {
			l.Status = model.LeadNew
			l.ConvertedAccountID, l.ConvertedContactID = nil, nil
		},
		keepManaged: func(in, stored *model.Lead) {
			in.Status = stored.Status
			in.ConvertedAccountID = stored.ConvertedAccountID
			in.ConvertedContactID = stored.ConvertedContactID
		},
		created: func(ctx context.Context, l *model.Lead) {
			notify(ctx, n, "New lead", "Lead #%d %s %s was created", l.ID, l.FirstName, l.Surname)
		},
	})
}

func NewCaseService(repo port.Repository[model.Case], n port.Notifier) port.Usecase[model.Case] {
	return newCRUD(repo, hooks[model.Case]{
		prepare: func(c *model.Case) {
			c.Status = model.CaseOpen
			c.ClosedAt = nil
			if c.Priority == "" {
				c.Priority = "medium"
			}
		},
		keepManaged: func(in, stored *model.Case) {
			in.Status = stored.Status
			in.ClosedAt = stored.ClosedAt
		},
		created: func(ctx context.Context, c *model.Case) {
			notify(ctx, n, "New case", "Case #%d was opened for contact #%d", c.ID, c.ContactID)
		},
	})
}

func NewContractService(repo port.Repository[model.Contract], orch port.ContractOrchestrator) port.Usecase[model.Contract] {
	return newCRUD(repo, hooks[model.Contract]{
		prepare: func(c *model.Contract) {
			c.Status = model.ContractDraft
			c.Approval = ""
			c.CopySignatureFrom(&model.Contract{})
		},
		keepManaged: func(in, stored *model.Contract) {
			in.Status = stored.Status
			in.Approval = stored.Approval
			in.CopySignatureFrom(stored)
		},
		created: func(ctx context.Context, c *model.Contract) { syncLifecycle(ctx, orch, c.ID) },
		updated: func(ctx context.Context, c *model.Contract) { syncLifecycle(ctx, orch, c.ID) },
		deleted: func(ctx context.Context, id uint) { syncLifecycle(ctx, orch, id) },
	})
}

func NewOpportunityService(repo port.Repository[model.Opportunity]) port.Usecase[model.Opportunity] {
	return newCRUD(repo, hooks[model.Opportunity]{
		prepare: func(o *model.Opportunity) { o.Stage = model.StageProspecting },
		keepManaged: func(in, stored *model.Opportunity) {
			in.Stage = stored.Stage
		},
	})
}
