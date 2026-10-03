package airflow_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "airflow-service"

const DefaultBaseURL = "http://airflow-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Health(ctx context.Context) (*model.AirflowHealthStatus, error) {
	path := "/health"
	q := url.Values{}
	h := http.Header{}
	var out *model.AirflowHealthStatus
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Dags(ctx context.Context, limit *int, offset *int) (*model.AirflowDAGList, error) {
	path := "/dags"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.AirflowDAGList
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Dag(ctx context.Context, dagID string) (*model.AirflowDag, error) {
	path := "/dags/" + url.PathEscape(dagID)
	q := url.Values{}
	h := http.Header{}
	var out *model.AirflowDag
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) DagRuns(ctx context.Context, dagID string, limit *int, offset *int) (*model.AirflowDAGRunList, error) {
	path := "/dags/" + url.PathEscape(dagID) + "/dagRuns"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.AirflowDAGRunList
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) DagRun(ctx context.Context, dagID string, dagRunID string) (*model.AirflowDAGRun, error) {
	path := "/dags/" + url.PathEscape(dagID) + "/dagRuns/" + url.PathEscape(dagRunID)
	q := url.Values{}
	h := http.Header{}
	var out *model.AirflowDAGRun
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) TaskInstances(ctx context.Context, dagID string, dagRunID string) (*model.AirflowTaskInstanceList, error) {
	path := "/dags/" + url.PathEscape(dagID) + "/dagRuns/" + url.PathEscape(dagRunID) + "/taskInstances"
	q := url.Values{}
	h := http.Header{}
	var out *model.AirflowTaskInstanceList
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) TaskInstance(ctx context.Context, dagID string, dagRunID string, taskID string) (*model.AirflowTaskInstance, error) {
	path := "/dags/" + url.PathEscape(dagID) + "/dagRuns/" + url.PathEscape(dagRunID) + "/taskInstances/" + url.PathEscape(taskID)
	q := url.Values{}
	h := http.Header{}
	var out *model.AirflowTaskInstance
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Variables(ctx context.Context, limit *int, offset *int) (*model.AirflowVariableList, error) {
	path := "/variables"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.AirflowVariableList
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Variable(ctx context.Context, key string) (*model.AirflowVariable, error) {
	path := "/variables/" + url.PathEscape(key)
	q := url.Values{}
	h := http.Header{}
	var out *model.AirflowVariable
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Pools(ctx context.Context, limit *int, offset *int) (*model.AirflowPoolList, error) {
	path := "/pools"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.AirflowPoolList
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Pool(ctx context.Context, name string) (*model.AirflowPool, error) {
	path := "/pools/" + url.PathEscape(name)
	q := url.Values{}
	h := http.Header{}
	var out *model.AirflowPool
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
