package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) CrmQuery() generated.CrmQueryResolver { return &crmQueryResolver{r} }

type crmQueryResolver struct{ *Resolver }

func (r *crmQueryResolver) Campaigns(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string) ([]*model.CrmCampaign, error) {
	return r.Clients.CustomerRelationshipService.Campaigns(ctx, offset, limit, qArg)
}

func (r *crmQueryResolver) Campaign(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmCampaign, error) {
	return r.Clients.CustomerRelationshipService.Campaign(ctx, id)
}

func (r *crmQueryResolver) Leads(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, status *string, source *string) ([]*model.CrmLead, error) {
	return r.Clients.CustomerRelationshipService.Leads(ctx, offset, limit, qArg, status, source)
}

func (r *crmQueryResolver) Lead(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmLead, error) {
	return r.Clients.CustomerRelationshipService.Lead(ctx, id)
}

func (r *crmQueryResolver) Accounts(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string) ([]*model.CrmAccount, error) {
	return r.Clients.CustomerRelationshipService.Accounts(ctx, offset, limit, qArg)
}

func (r *crmQueryResolver) Account(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmAccount, error) {
	return r.Clients.CustomerRelationshipService.Account(ctx, id)
}

func (r *crmQueryResolver) Contacts(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, accountID *string) ([]*model.CrmContact, error) {
	return r.Clients.CustomerRelationshipService.Contacts(ctx, offset, limit, qArg, accountID)
}

func (r *crmQueryResolver) Contact(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmContact, error) {
	return r.Clients.CustomerRelationshipService.Contact(ctx, id)
}

func (r *crmQueryResolver) CampaignMembers(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, campaignID *string, leadID *string, contactID *string) ([]*model.CrmCampaignMember, error) {
	return r.Clients.CustomerRelationshipService.CampaignMembers(ctx, offset, limit, qArg, campaignID, leadID, contactID)
}

func (r *crmQueryResolver) CampaignMember(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmCampaignMember, error) {
	return r.Clients.CustomerRelationshipService.CampaignMember(ctx, id)
}

func (r *crmQueryResolver) Cases(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, contactID *string, status *string, priority *string) ([]*model.CrmCase, error) {
	return r.Clients.CustomerRelationshipService.Cases(ctx, offset, limit, qArg, contactID, status, priority)
}

func (r *crmQueryResolver) Case(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmCase, error) {
	return r.Clients.CustomerRelationshipService.Case(ctx, id)
}

func (r *crmQueryResolver) Contracts(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, accountID *string, contractStatus *string) ([]*model.CrmContract, error) {
	return r.Clients.CustomerRelationshipService.Contracts(ctx, offset, limit, qArg, accountID, contractStatus)
}

func (r *crmQueryResolver) Contract(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmContract, error) {
	return r.Clients.CustomerRelationshipService.Contract(ctx, id)
}

func (r *crmQueryResolver) AccountContactRoles(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, contactID *string, accountID *string) ([]*model.CrmAccountContactRole, error) {
	return r.Clients.CustomerRelationshipService.AccountContactRoles(ctx, offset, limit, qArg, contactID, accountID)
}

func (r *crmQueryResolver) AccountContactRole(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmAccountContactRole, error) {
	return r.Clients.CustomerRelationshipService.AccountContactRole(ctx, id)
}

func (r *crmQueryResolver) Opportunities(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, accountID *string, opportunityStage *string) ([]*model.CrmOpportunity, error) {
	return r.Clients.CustomerRelationshipService.Opportunities(ctx, offset, limit, qArg, accountID, opportunityStage)
}

func (r *crmQueryResolver) Opportunity(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmOpportunity, error) {
	return r.Clients.CustomerRelationshipService.Opportunity(ctx, id)
}

func (r *crmQueryResolver) OpportunityContactRoles(ctx context.Context, obj *model.CrmQuery, offset *int, limit *int, qArg *string, contactID *string, opportunityID *string) ([]*model.CrmOpportunityContactRole, error) {
	return r.Clients.CustomerRelationshipService.OpportunityContactRoles(ctx, offset, limit, qArg, contactID, opportunityID)
}

func (r *crmQueryResolver) OpportunityContactRole(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmOpportunityContactRole, error) {
	return r.Clients.CustomerRelationshipService.OpportunityContactRole(ctx, id)
}

func (r *crmQueryResolver) ContractFileKinds(ctx context.Context, obj *model.CrmQuery) ([]*model.CrmFileKind, error) {
	return r.Clients.CustomerRelationshipService.ContractFileKinds(ctx)
}

func (r *crmQueryResolver) ContractFiles(ctx context.Context, obj *model.CrmQuery, id int, kind *string, allVersions *bool, offset *int, limit *int) ([]*model.CrmContractFile, error) {
	return r.Clients.CustomerRelationshipService.ContractFiles(ctx, id, kind, allVersions, offset, limit)
}

func (r *crmQueryResolver) ContractFile(ctx context.Context, obj *model.CrmQuery, id int, fileID int) (*model.CrmContractFile, error) {
	return r.Clients.CustomerRelationshipService.ContractFile(ctx, id, fileID)
}

func (r *crmQueryResolver) ContractFileHistory(ctx context.Context, obj *model.CrmQuery, id int, fileID int) (*model.CrmFileHistory, error) {
	return r.Clients.CustomerRelationshipService.ContractFileHistory(ctx, id, fileID)
}

func (r *crmQueryResolver) ContractFileEvents(ctx context.Context, obj *model.CrmQuery, id int, action *string, fileID *int, offset *int, limit *int) ([]*model.CrmContractFileEvent, error) {
	return r.Clients.CustomerRelationshipService.ContractFileEvents(ctx, id, action, fileID, offset, limit)
}

func (r *crmQueryResolver) ContractSigningPayload(ctx context.Context, obj *model.CrmQuery, id int, signedBy string, endDate string, signingTime *string) (*model.CrmSigningPayload, error) {
	return r.Clients.CustomerRelationshipService.ContractSigningPayload(ctx, id, signedBy, endDate, signingTime)
}

func (r *crmQueryResolver) AccountOverview(ctx context.Context, obj *model.CrmQuery, id int) (*model.CrmAccountOverview, error) {
	return r.Clients.CustomerRelationshipService.AccountOverview(ctx, id)
}

func (r *crmQueryResolver) ReportSummary(ctx context.Context, obj *model.CrmQuery) (*model.CrmSummary, error) {
	return r.Clients.CustomerRelationshipService.ReportSummary(ctx)
}
