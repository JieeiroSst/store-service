package toggle_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "toggle-service"

const DefaultBaseURL = "http://toggle-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Roles(ctx context.Context) ([]*model.ToggleRole, error) {
	path := "/api/admin/roles"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToggleRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Environments(ctx context.Context) ([]*model.ToggleEnvironment, error) {
	path := "/api/admin/environments/"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToggleEnvironment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Tokens(ctx context.Context) ([]*model.ToggleAPIToken, error) {
	path := "/api/admin/tokens/"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToggleAPIToken
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Projects(ctx context.Context) ([]*model.ToggleProject, error) {
	path := "/api/admin/projects/"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToggleProject
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Project(ctx context.Context, projectID string) (*model.ToggleProject, error) {
	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/"
	q := url.Values{}
	h := http.Header{}
	var out *model.ToggleProject
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ProjectMembers(ctx context.Context, projectID string) ([]*model.ToggleProjectMembership, error) {
	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/members"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToggleProjectMembership
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Audit(ctx context.Context, projectID string, entityType *string, since *string, until *string) ([]*model.ToggleAuditEvent, error) {
	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/audit"
	q := url.Values{}
	if entityType != nil {
		q.Set("entityType", *entityType)
	}
	if since != nil {
		q.Set("since", *since)
	}
	if until != nil {
		q.Set("until", *until)
	}
	h := http.Header{}
	var out []*model.ToggleAuditEvent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Flags(ctx context.Context, projectID string) ([]*model.ToggleFeatureFlag, error) {
	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/flags"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToggleFeatureFlag
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Flag(ctx context.Context, projectID string, key string) (*model.ToggleFeatureFlag, error) {
	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/flags/" + url.PathEscape(key) + "/"
	q := url.Values{}
	h := http.Header{}
	var out *model.ToggleFeatureFlag
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Strategies(ctx context.Context, projectID string, key string, envName string) ([]*model.ToggleActivationStrategy, error) {
	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/flags/" + url.PathEscape(key) + "/environments/" + url.PathEscape(envName) + "/strategies"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToggleActivationStrategy
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientFeatures(ctx context.Context) (*model.ToggleClientFeaturesResponse, error) {
	path := "/api/client/features"
	q := url.Values{}
	h := http.Header{}
	var out *model.ToggleClientFeaturesResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
