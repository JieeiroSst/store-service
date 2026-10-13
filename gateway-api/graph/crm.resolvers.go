package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CrmCreateCampaign(ctx context.Context, input model.CrmCampaignInput) (*model.CrmCampaign, error) {
	return restCall[*model.CrmCampaign](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/campaigns", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateCampaign(ctx context.Context, id int, input model.CrmCampaignInput) (*model.CrmCampaign, error) {
	return restCall[*model.CrmCampaign](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/campaigns/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteCampaign(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/campaigns/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateLead(ctx context.Context, input model.CrmLeadInput) (*model.CrmLead, error) {
	return restCall[*model.CrmLead](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/leads", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateLead(ctx context.Context, id int, input model.CrmLeadInput) (*model.CrmLead, error) {
	return restCall[*model.CrmLead](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/leads/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteLead(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/leads/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateAccount(ctx context.Context, input model.CrmAccountInput) (*model.CrmAccount, error) {
	return restCall[*model.CrmAccount](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/accounts", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateAccount(ctx context.Context, id int, input model.CrmAccountInput) (*model.CrmAccount, error) {
	return restCall[*model.CrmAccount](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/accounts/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteAccount(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/accounts/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateContact(ctx context.Context, input model.CrmContactInput) (*model.CrmContact, error) {
	return restCall[*model.CrmContact](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/contacts", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateContact(ctx context.Context, id int, input model.CrmContactInput) (*model.CrmContact, error) {
	return restCall[*model.CrmContact](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/contacts/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteContact(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/contacts/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateCampaignMember(ctx context.Context, input model.CrmCampaignMemberInput) (*model.CrmCampaignMember, error) {
	return restCall[*model.CrmCampaignMember](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/campaign-members", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateCampaignMember(ctx context.Context, id int, input model.CrmCampaignMemberInput) (*model.CrmCampaignMember, error) {
	return restCall[*model.CrmCampaignMember](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/campaign-members/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteCampaignMember(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/campaign-members/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateCase(ctx context.Context, input model.CrmCaseInput) (*model.CrmCase, error) {
	return restCall[*model.CrmCase](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/cases", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateCase(ctx context.Context, id int, input model.CrmCaseInput) (*model.CrmCase, error) {
	return restCall[*model.CrmCase](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/cases/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteCase(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/cases/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateContract(ctx context.Context, input model.CrmContractInput) (*model.CrmContract, error) {
	return restCall[*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/contracts", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateContract(ctx context.Context, id int, input model.CrmContractInput) (*model.CrmContract, error) {
	return restCall[*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/contracts/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteContract(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/contracts/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateAccountContactRole(ctx context.Context, input model.CrmAccountContactRoleInput) (*model.CrmAccountContactRole, error) {
	return restCall[*model.CrmAccountContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/account-contact-roles", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateAccountContactRole(ctx context.Context, id int, input model.CrmAccountContactRoleInput) (*model.CrmAccountContactRole, error) {
	return restCall[*model.CrmAccountContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/account-contact-roles/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteAccountContactRole(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/account-contact-roles/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateOpportunity(ctx context.Context, input model.CrmOpportunityInput) (*model.CrmOpportunity, error) {
	return restCall[*model.CrmOpportunity](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/opportunities", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateOpportunity(ctx context.Context, id int, input model.CrmOpportunityInput) (*model.CrmOpportunity, error) {
	return restCall[*model.CrmOpportunity](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/opportunities/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteOpportunity(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/opportunities/%v", id)}, "")
}

func (r *mutationResolver) CrmCreateOpportunityContactRole(ctx context.Context, input model.CrmOpportunityContactRoleInput) (*model.CrmOpportunityContactRole, error) {
	return restCall[*model.CrmOpportunityContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: "/api/v1/opportunity-contact-roles", Body: input}, "")
}

func (r *mutationResolver) CrmUpdateOpportunityContactRole(ctx context.Context, id int, input model.CrmOpportunityContactRoleInput) (*model.CrmOpportunityContactRole, error) {
	return restCall[*model.CrmOpportunityContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/opportunity-contact-roles/%v", id), Body: input}, "")
}

func (r *mutationResolver) CrmDeleteOpportunityContactRole(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/opportunity-contact-roles/%v", id)}, "")
}

func (r *mutationResolver) CrmConvertLead(ctx context.Context, id int, input *model.CrmConvertLeadInput) (*model.CrmConvertLeadResult, error) {
	return restCall[*model.CrmConvertLeadResult](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/leads/%v/convert", id), Body: input}, "")
}

func (r *mutationResolver) CrmSetOpportunityStage(ctx context.Context, id int, stage string) (*model.CrmOpportunity, error) {
	return restCall[*model.CrmOpportunity](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/opportunities/%v/stage", id), Body: map[string]any{"stage": stage}}, "")
}

func (r *mutationResolver) CrmSignContract(ctx context.Context, id int, input model.CrmSignContractInput) (*model.CrmContract, error) {
	return restCall[*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/contracts/%v/sign", id), Body: input}, "")
}

func (r *mutationResolver) CrmSubmitContract(ctx context.Context, id int) (*model.CrmContract, error) {
	return restCall[*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/contracts/%v/submit", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CrmApproveContract(ctx context.Context, id int) (*model.CrmContract, error) {
	return restCall[*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/contracts/%v/approve", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CrmRejectContract(ctx context.Context, id int) (*model.CrmContract, error) {
	return restCall[*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/contracts/%v/reject", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CrmCloseCase(ctx context.Context, id int) (*model.CrmCase, error) {
	return restCall[*model.CrmCase](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cases/%v/close", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CrmUploadContractFile(ctx context.Context, id int, file graphql.Upload, kind *string, description *string, replaces *int) (*model.CrmContractFile, error) {
	return restMultipart[*model.CrmContractFile](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/contracts/%v/files", id)}, map[string]any{"kind": kind, "description": description, "replaces": replaces}, uploadFiles("file", file), "")
}

func (r *mutationResolver) CrmDeleteContractFile(ctx context.Context, id int, fileID int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/contracts/%v/files/%v", id, fileID)}, "")
}

func (r *queryResolver) CrmCampaigns(ctx context.Context, offset *int, limit *int, q *string) ([]*model.CrmCampaign, error) {
	return restCall[[]*model.CrmCampaign](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/campaigns", Query: restQuery("offset", offset, "limit", limit, "q", q)}, "")
}

func (r *queryResolver) CrmCampaign(ctx context.Context, id int) (*model.CrmCampaign, error) {
	return restCall[*model.CrmCampaign](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/campaigns/%v", id)}, "")
}

func (r *queryResolver) CrmLeads(ctx context.Context, offset *int, limit *int, q *string, status *string, source *string) ([]*model.CrmLead, error) {
	return restCall[[]*model.CrmLead](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/leads", Query: restQuery("offset", offset, "limit", limit, "q", q, "status", status, "source", source)}, "")
}

func (r *queryResolver) CrmLead(ctx context.Context, id int) (*model.CrmLead, error) {
	return restCall[*model.CrmLead](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/leads/%v", id)}, "")
}

func (r *queryResolver) CrmAccounts(ctx context.Context, offset *int, limit *int, q *string) ([]*model.CrmAccount, error) {
	return restCall[[]*model.CrmAccount](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/accounts", Query: restQuery("offset", offset, "limit", limit, "q", q)}, "")
}

func (r *queryResolver) CrmAccount(ctx context.Context, id int) (*model.CrmAccount, error) {
	return restCall[*model.CrmAccount](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/accounts/%v", id)}, "")
}

func (r *queryResolver) CrmContacts(ctx context.Context, offset *int, limit *int, q *string, accountID *int) ([]*model.CrmContact, error) {
	return restCall[[]*model.CrmContact](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/contacts", Query: restQuery("offset", offset, "limit", limit, "q", q, "account_id", accountID)}, "")
}

func (r *queryResolver) CrmContact(ctx context.Context, id int) (*model.CrmContact, error) {
	return restCall[*model.CrmContact](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/contacts/%v", id)}, "")
}

func (r *queryResolver) CrmCampaignMembers(ctx context.Context, offset *int, limit *int, q *string, campaignID *int, leadID *int, contactID *int) ([]*model.CrmCampaignMember, error) {
	return restCall[[]*model.CrmCampaignMember](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/campaign-members", Query: restQuery("offset", offset, "limit", limit, "q", q, "campaign_id", campaignID, "lead_id", leadID, "contact_id", contactID)}, "")
}

func (r *queryResolver) CrmCampaignMember(ctx context.Context, id int) (*model.CrmCampaignMember, error) {
	return restCall[*model.CrmCampaignMember](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/campaign-members/%v", id)}, "")
}

func (r *queryResolver) CrmCases(ctx context.Context, offset *int, limit *int, q *string, contactID *int, status *string, priority *string) ([]*model.CrmCase, error) {
	return restCall[[]*model.CrmCase](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/cases", Query: restQuery("offset", offset, "limit", limit, "q", q, "contact_id", contactID, "status", status, "priority", priority)}, "")
}

func (r *queryResolver) CrmCase(ctx context.Context, id int) (*model.CrmCase, error) {
	return restCall[*model.CrmCase](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/cases/%v", id)}, "")
}

func (r *queryResolver) CrmContracts(ctx context.Context, offset *int, limit *int, q *string, accountID *int, contractStatus *string) ([]*model.CrmContract, error) {
	return restCall[[]*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/contracts", Query: restQuery("offset", offset, "limit", limit, "q", q, "account_id", accountID, "contract_status", contractStatus)}, "")
}

func (r *queryResolver) CrmContract(ctx context.Context, id int) (*model.CrmContract, error) {
	return restCall[*model.CrmContract](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/contracts/%v", id)}, "")
}

func (r *queryResolver) CrmAccountContactRoles(ctx context.Context, offset *int, limit *int, q *string, contactID *int, accountID *int) ([]*model.CrmAccountContactRole, error) {
	return restCall[[]*model.CrmAccountContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/account-contact-roles", Query: restQuery("offset", offset, "limit", limit, "q", q, "contact_id", contactID, "account_id", accountID)}, "")
}

func (r *queryResolver) CrmAccountContactRole(ctx context.Context, id int) (*model.CrmAccountContactRole, error) {
	return restCall[*model.CrmAccountContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/account-contact-roles/%v", id)}, "")
}

func (r *queryResolver) CrmOpportunities(ctx context.Context, offset *int, limit *int, q *string, accountID *int, opportunityStage *string) ([]*model.CrmOpportunity, error) {
	return restCall[[]*model.CrmOpportunity](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/opportunities", Query: restQuery("offset", offset, "limit", limit, "q", q, "account_id", accountID, "opportunity_stage", opportunityStage)}, "")
}

func (r *queryResolver) CrmOpportunity(ctx context.Context, id int) (*model.CrmOpportunity, error) {
	return restCall[*model.CrmOpportunity](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/opportunities/%v", id)}, "")
}

func (r *queryResolver) CrmOpportunityContactRoles(ctx context.Context, offset *int, limit *int, q *string, contactID *int, opportunityID *int) ([]*model.CrmOpportunityContactRole, error) {
	return restCall[[]*model.CrmOpportunityContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/opportunity-contact-roles", Query: restQuery("offset", offset, "limit", limit, "q", q, "contact_id", contactID, "opportunity_id", opportunityID)}, "")
}

func (r *queryResolver) CrmOpportunityContactRole(ctx context.Context, id int) (*model.CrmOpportunityContactRole, error) {
	return restCall[*model.CrmOpportunityContactRole](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/opportunity-contact-roles/%v", id)}, "")
}

func (r *queryResolver) CrmContractSigningPayload(ctx context.Context, id int, endDate string, signedBy string, signingTime *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/contracts/%v/signing-payload", id), Query: restQuery("end_date", endDate, "signed_by", signedBy, "signing_time", signingTime)}, "")
}

func (r *queryResolver) CrmAccountOverview(ctx context.Context, id int) (*model.CrmAccountOverview, error) {
	return restCall[*model.CrmAccountOverview](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/accounts/%v/overview", id)}, "")
}

func (r *queryResolver) CrmSummary(ctx context.Context) (*model.CrmSummary, error) {
	return restCall[*model.CrmSummary](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/reports/summary"}, "")
}

func (r *queryResolver) CrmContractFileKinds(ctx context.Context) ([]*model.CrmFileKind, error) {
	return restCall[[]*model.CrmFileKind](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: "/api/v1/contract-file-kinds"}, "")
}

func (r *queryResolver) CrmContractFiles(ctx context.Context, id int, kind *string, allVersions *bool, offset *int, limit *int) ([]*model.CrmContractFile, error) {
	return restCall[[]*model.CrmContractFile](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/contracts/%v/files", id), Query: restQuery("kind", kind, "all_versions", allVersions, "offset", offset, "limit", limit)}, "")
}

func (r *queryResolver) CrmContractFile(ctx context.Context, id int, fileID int) (*model.CrmContractFile, error) {
	return restCall[*model.CrmContractFile](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/contracts/%v/files/%v", id, fileID)}, "")
}

func (r *queryResolver) CrmContractFileHistory(ctx context.Context, id int, fileID int) (*model.CrmFileHistory, error) {
	return restCall[*model.CrmFileHistory](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/contracts/%v/files/%v/history", id, fileID)}, "")
}

func (r *queryResolver) CrmContractFileEvents(ctx context.Context, id int, action *string, fileID *int, offset *int, limit *int) ([]*model.CrmContractFileEvent, error) {
	return restCall[[]*model.CrmContractFileEvent](ctx, r.Proxy, "customer-relationship-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/contracts/%v/file-events", id), Query: restQuery("action", action, "file_id", fileID, "offset", offset, "limit", limit)}, "")
}
