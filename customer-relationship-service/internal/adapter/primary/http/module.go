package http

import (
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"go.uber.org/fx"
)

func resource[T any, PT model.Entity[T]](path string, filters ...string) any {
	return fx.Annotate(
		func(uc port.Usecase[T]) *crudHandler[T, PT] {
			return &crudHandler[T, PT]{path: path, uc: uc, filters: filters}
		},
		fx.As(new(Resource)),
		fx.ResultTags(`group:"resources"`),
	)
}

func extra(constructor any) any {
	return fx.Annotate(constructor, fx.As(new(Resource)), fx.ResultTags(`group:"resources"`))
}

var Module = fx.Options(
	fx.Provide(
		resource[model.Campaign]("/campaigns"),
		resource[model.Lead]("/leads", "status", "source"),
		resource[model.Account]("/accounts"),
		resource[model.Contact]("/contacts", "account_id"),
		resource[model.CampaignMember]("/campaign-members", "campaign_id", "lead_id", "contact_id"),
		resource[model.Case]("/cases", "contact_id", "status", "priority"),
		resource[model.Contract]("/contracts", "account_id", "contract_status"),
		resource[model.AccountContactRole]("/account-contact-roles", "contact_id", "account_id"),
		resource[model.Opportunity]("/opportunities", "account_id", "opportunity_stage"),
		resource[model.OpportunityContactRole]("/opportunity-contact-roles", "contact_id", "opportunity_id"),

		extra(NewLeadHandler),
		extra(NewOpportunityHandler),
		extra(NewContractHandler),
		extra(NewCaseHandler),
		extra(NewReportHandler),
		extra(NewFileHandler),
	),
	fx.Provide(NewAuthenticator, NewRouter),
)
