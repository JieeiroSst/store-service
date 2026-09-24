package application

import (
	"context"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type reportService struct{ reporter port.Reporter }

func NewReportService(r port.Reporter) port.ReportUsecase { return &reportService{reporter: r} }

func (s *reportService) Summary(ctx context.Context) (*model.Summary, error) {
	return s.reporter.Summary(ctx)
}

type accountOverview struct {
	accounts  port.Repository[model.Account]
	contacts  port.Repository[model.Contact]
	contracts port.Repository[model.Contract]
	opps      port.Repository[model.Opportunity]
}

func NewAccountOverview(
	accounts port.Repository[model.Account],
	contacts port.Repository[model.Contact],
	contracts port.Repository[model.Contract],
	opps port.Repository[model.Opportunity],
) port.AccountOverviewUsecase {
	return &accountOverview{accounts: accounts, contacts: contacts, contracts: contracts, opps: opps}
}

const overviewLimit = 200

func (s *accountOverview) Overview(ctx context.Context, id uint) (*port.AccountOverview, error) {
	account, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	q := port.ListQuery{Limit: overviewLimit, Equals: map[string]any{"account_id": id}}

	contacts, _, err := s.contacts.List(ctx, q)
	if err != nil {
		return nil, err
	}
	contracts, _, err := s.contracts.List(ctx, q)
	if err != nil {
		return nil, err
	}
	opps, _, err := s.opps.List(ctx, q)
	if err != nil {
		return nil, err
	}
	return &port.AccountOverview{Account: account, Contacts: contacts, Contracts: contracts, Opportunities: opps}, nil
}
