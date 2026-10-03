package customer_relationship_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "customer-relationship-service"

const DefaultBaseURL = "http://customer-relationship-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Campaigns(ctx context.Context, offset *int, limit *int, qArg *string) ([]*model.CrmCampaign, error) {
	path := "/api/v1/campaigns"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	h := http.Header{}
	var out []*model.CrmCampaign
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Campaign(ctx context.Context, id int) (*model.CrmCampaign, error) {
	path := "/api/v1/campaigns/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmCampaign
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Leads(ctx context.Context, offset *int, limit *int, qArg *string, status *string, source *string) ([]*model.CrmLead, error) {
	path := "/api/v1/leads"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if source != nil {
		q.Set("source", *source)
	}
	h := http.Header{}
	var out []*model.CrmLead
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Lead(ctx context.Context, id int) (*model.CrmLead, error) {
	path := "/api/v1/leads/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmLead
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Accounts(ctx context.Context, offset *int, limit *int, qArg *string) ([]*model.CrmAccount, error) {
	path := "/api/v1/accounts"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	h := http.Header{}
	var out []*model.CrmAccount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Account(ctx context.Context, id int) (*model.CrmAccount, error) {
	path := "/api/v1/accounts/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmAccount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Contacts(ctx context.Context, offset *int, limit *int, qArg *string, accountID *string) ([]*model.CrmContact, error) {
	path := "/api/v1/contacts"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if accountID != nil {
		q.Set("account_id", *accountID)
	}
	h := http.Header{}
	var out []*model.CrmContact
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Contact(ctx context.Context, id int) (*model.CrmContact, error) {
	path := "/api/v1/contacts/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmContact
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CampaignMembers(ctx context.Context, offset *int, limit *int, qArg *string, campaignID *string, leadID *string, contactID *string) ([]*model.CrmCampaignMember, error) {
	path := "/api/v1/campaign-members"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if campaignID != nil {
		q.Set("campaign_id", *campaignID)
	}
	if leadID != nil {
		q.Set("lead_id", *leadID)
	}
	if contactID != nil {
		q.Set("contact_id", *contactID)
	}
	h := http.Header{}
	var out []*model.CrmCampaignMember
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CampaignMember(ctx context.Context, id int) (*model.CrmCampaignMember, error) {
	path := "/api/v1/campaign-members/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmCampaignMember
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Cases(ctx context.Context, offset *int, limit *int, qArg *string, contactID *string, status *string, priority *string) ([]*model.CrmCase, error) {
	path := "/api/v1/cases"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if contactID != nil {
		q.Set("contact_id", *contactID)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if priority != nil {
		q.Set("priority", *priority)
	}
	h := http.Header{}
	var out []*model.CrmCase
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Case(ctx context.Context, id int) (*model.CrmCase, error) {
	path := "/api/v1/cases/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmCase
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Contracts(ctx context.Context, offset *int, limit *int, qArg *string, accountID *string, contractStatus *string) ([]*model.CrmContract, error) {
	path := "/api/v1/contracts"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if accountID != nil {
		q.Set("account_id", *accountID)
	}
	if contractStatus != nil {
		q.Set("contract_status", *contractStatus)
	}
	h := http.Header{}
	var out []*model.CrmContract
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Contract(ctx context.Context, id int) (*model.CrmContract, error) {
	path := "/api/v1/contracts/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmContract
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AccountContactRoles(ctx context.Context, offset *int, limit *int, qArg *string, contactID *string, accountID *string) ([]*model.CrmAccountContactRole, error) {
	path := "/api/v1/account-contact-roles"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if contactID != nil {
		q.Set("contact_id", *contactID)
	}
	if accountID != nil {
		q.Set("account_id", *accountID)
	}
	h := http.Header{}
	var out []*model.CrmAccountContactRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AccountContactRole(ctx context.Context, id int) (*model.CrmAccountContactRole, error) {
	path := "/api/v1/account-contact-roles/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmAccountContactRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Opportunities(ctx context.Context, offset *int, limit *int, qArg *string, accountID *string, opportunityStage *string) ([]*model.CrmOpportunity, error) {
	path := "/api/v1/opportunities"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if accountID != nil {
		q.Set("account_id", *accountID)
	}
	if opportunityStage != nil {
		q.Set("opportunity_stage", *opportunityStage)
	}
	h := http.Header{}
	var out []*model.CrmOpportunity
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Opportunity(ctx context.Context, id int) (*model.CrmOpportunity, error) {
	path := "/api/v1/opportunities/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmOpportunity
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) OpportunityContactRoles(ctx context.Context, offset *int, limit *int, qArg *string, contactID *string, opportunityID *string) ([]*model.CrmOpportunityContactRole, error) {
	path := "/api/v1/opportunity-contact-roles"
	q := url.Values{}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if contactID != nil {
		q.Set("contact_id", *contactID)
	}
	if opportunityID != nil {
		q.Set("opportunity_id", *opportunityID)
	}
	h := http.Header{}
	var out []*model.CrmOpportunityContactRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) OpportunityContactRole(ctx context.Context, id int) (*model.CrmOpportunityContactRole, error) {
	path := "/api/v1/opportunity-contact-roles/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmOpportunityContactRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ContractFileKinds(ctx context.Context) ([]*model.CrmFileKind, error) {
	path := "/api/v1/contract-file-kinds"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CrmFileKind
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ContractFiles(ctx context.Context, id int, kind *string, allVersions *bool, offset *int, limit *int) ([]*model.CrmContractFile, error) {
	path := "/api/v1/contracts/" + url.PathEscape(strconv.Itoa(id)) + "/files"
	q := url.Values{}
	if kind != nil {
		q.Set("kind", *kind)
	}
	if allVersions != nil {
		q.Set("all_versions", strconv.FormatBool(*allVersions))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.CrmContractFile
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ContractFile(ctx context.Context, id int, fileID int) (*model.CrmContractFile, error) {
	path := "/api/v1/contracts/" + url.PathEscape(strconv.Itoa(id)) + "/files/" + url.PathEscape(strconv.Itoa(fileID))
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmContractFile
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ContractFileHistory(ctx context.Context, id int, fileID int) (*model.CrmFileHistory, error) {
	path := "/api/v1/contracts/" + url.PathEscape(strconv.Itoa(id)) + "/files/" + url.PathEscape(strconv.Itoa(fileID)) + "/history"
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmFileHistory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ContractFileEvents(ctx context.Context, id int, action *string, fileID *int, offset *int, limit *int) ([]*model.CrmContractFileEvent, error) {
	path := "/api/v1/contracts/" + url.PathEscape(strconv.Itoa(id)) + "/file-events"
	q := url.Values{}
	if action != nil {
		q.Set("action", *action)
	}
	if fileID != nil {
		q.Set("file_id", strconv.Itoa(*fileID))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.CrmContractFileEvent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ContractSigningPayload(ctx context.Context, id int, signedBy string, endDate string, signingTime *string) (*model.CrmSigningPayload, error) {
	path := "/api/v1/contracts/" + url.PathEscape(strconv.Itoa(id)) + "/signing-payload"
	q := url.Values{}
	q.Set("signed_by", signedBy)
	q.Set("end_date", endDate)
	if signingTime != nil {
		q.Set("signing_time", *signingTime)
	}
	h := http.Header{}
	var out *model.CrmSigningPayload
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AccountOverview(ctx context.Context, id int) (*model.CrmAccountOverview, error) {
	path := "/api/v1/accounts/" + url.PathEscape(strconv.Itoa(id)) + "/overview"
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmAccountOverview
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ReportSummary(ctx context.Context) (*model.CrmSummary, error) {
	path := "/api/v1/reports/summary"
	q := url.Values{}
	h := http.Header{}
	var out *model.CrmSummary
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
