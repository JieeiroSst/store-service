package repository

import (
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		New[model.Campaign],
		New[model.Lead],
		New[model.Account],
		New[model.Contact],
		New[model.CampaignMember],
		New[model.Case],
		New[model.Contract],
		New[model.AccountContactRole],
		New[model.Opportunity],
		New[model.OpportunityContactRole],
		NewFileRepository,
		New[model.ContractFileEvent],
		NewTxRunner,
		NewReporter,
		NewContractExpirer,
	),
)
