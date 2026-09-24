package application

import (
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewLeadService,
		NewCaseService,
		NewContractService,
		NewOpportunityService,
		NewCRUDService[model.Campaign],
		NewCRUDService[model.Account],
		NewCRUDService[model.Contact],
		NewCRUDService[model.CampaignMember],
		NewCRUDService[model.AccountContactRole],
		NewCRUDService[model.OpportunityContactRole],
	),
	fx.Provide(
		NewLeadConversion,
		NewOpportunityWorkflow,
		NewContractWorkflow,
		NewContractExpiry,
		NewContractLifecycle,
		NewContractFiles,
		func(f *ContractFiles) port.ContractFileUsecase { return f },
		func(f *ContractFiles) port.ContractFileKeeper { return f },
		NewCaseWorkflow,
		NewAccountOverview,
		NewReportService,
	),
)
