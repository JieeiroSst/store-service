package tool_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "tool-service"

const DefaultBaseURL = "http://tool-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Jobs(ctx context.Context) ([]*model.ToolJob, error) {
	path := "/api/v1/jobs"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToolJob
	err := c.rest.Get(ctx, path, q, h, "jobs", &out)
	return out, err
}

func (c *Client) Job(ctx context.Context, id string) (*model.ToolJob, error) {
	path := "/api/v1/jobs/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.ToolJob
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MobileDevices(ctx context.Context) (*model.ToolMobileDevices, error) {
	path := "/api/v1/qa/mobile/devices"
	q := url.Values{}
	h := http.Header{}
	var out *model.ToolMobileDevices
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) LearningStats(ctx context.Context) (*model.ToolStats, error) {
	path := "/api/v1/learning/stats"
	q := url.Values{}
	h := http.Header{}
	var out *model.ToolStats
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) LearningPending(ctx context.Context) ([]*model.ToolExperience, error) {
	path := "/api/v1/learning/pending"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToolExperience
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) LearningHistory(ctx context.Context) ([]*model.ToolSnapshot, error) {
	path := "/api/v1/learning/history"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ToolSnapshot
	err := c.rest.Get(ctx, path, q, h, "history", &out)
	return out, err
}

func (c *Client) Suite(ctx context.Context, name string) (*model.ToolSuiteResult, error) {
	path := "/api/v1/suites/" + url.PathEscape(name)
	q := url.Values{}
	h := http.Header{}
	var out *model.ToolSuiteResult
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
