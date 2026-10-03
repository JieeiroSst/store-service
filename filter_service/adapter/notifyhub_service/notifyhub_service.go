package notifyhub_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "notifyhub-service"

const DefaultBaseURL = "http://notifyhub-service:8095"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Channels(ctx context.Context, typeArg *string, active *bool) ([]*model.HubChannel, error) {
	path := "/api/v1/channels"
	q := url.Values{}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	if active != nil {
		q.Set("active", strconv.FormatBool(*active))
	}
	h := http.Header{}
	var out []*model.HubChannel
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Channel(ctx context.Context, id string) (*model.HubChannel, error) {
	path := "/api/v1/channels/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.HubChannel
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) DataSources(ctx context.Context) ([]*model.HubDataSource, error) {
	path := "/api/v1/data-sources"
	q := url.Values{}
	h := http.Header{}
	var out []*model.HubDataSource
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) DataSource(ctx context.Context, id string) (*model.HubDataSource, error) {
	path := "/api/v1/data-sources/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.HubDataSource
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Templates(ctx context.Context, channel *string) ([]*model.HubTemplate, error) {
	path := "/api/v1/templates"
	q := url.Values{}
	if channel != nil {
		q.Set("channel", *channel)
	}
	h := http.Header{}
	var out []*model.HubTemplate
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Template(ctx context.Context, id string) (*model.HubTemplate, error) {
	path := "/api/v1/templates/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.HubTemplate
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Jobs(ctx context.Context, status *string, page *int, pageSize *int) (*model.HubJobPage, error) {
	path := "/api/v1/jobs"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	h := http.Header{}
	var out *model.HubJobPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Job(ctx context.Context, id string) (*model.HubNotifyJob, error) {
	path := "/api/v1/jobs/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.HubNotifyJob
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) History(ctx context.Context, jobID *string, status *string, page *int, pageSize *int) (*model.HubHistoryPage, error) {
	path := "/api/v1/history"
	q := url.Values{}
	if jobID != nil {
		q.Set("job_id", *jobID)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	h := http.Header{}
	var out *model.HubHistoryPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SchedulerStatus(ctx context.Context) (*model.HubSchedulerStatus, error) {
	path := "/api/v1/scheduler/status"
	q := url.Values{}
	h := http.Header{}
	var out *model.HubSchedulerStatus
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
