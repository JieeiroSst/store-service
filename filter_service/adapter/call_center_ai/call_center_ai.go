package call_center_ai

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "call_center_ai"

const DefaultBaseURL = "http://call-center-ai:8000"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Calls(ctx context.Context, skip *int, limit *int, status *string, fromNumber *string) ([]*model.CallCenterCall, error) {
	path := "/api/calls"
	q := url.Values{}
	if skip != nil {
		q.Set("skip", strconv.Itoa(*skip))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if fromNumber != nil {
		q.Set("from_number", *fromNumber)
	}
	h := http.Header{}
	var out []*model.CallCenterCall
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Call(ctx context.Context, callID int) (*model.CallCenterCallHistory, error) {
	path := "/api/calls/" + url.PathEscape(strconv.Itoa(callID))
	q := url.Values{}
	h := http.Header{}
	var out *model.CallCenterCallHistory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Scenarios(ctx context.Context, skip *int, limit *int) ([]*model.CallCenterScenario, error) {
	path := "/api/scenarios"
	q := url.Values{}
	if skip != nil {
		q.Set("skip", strconv.Itoa(*skip))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.CallCenterScenario
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Scenario(ctx context.Context, scenarioID int) (*model.CallCenterScenario, error) {
	path := "/api/scenarios/" + url.PathEscape(strconv.Itoa(scenarioID))
	q := url.Values{}
	h := http.Header{}
	var out *model.CallCenterScenario
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Customer(ctx context.Context, phoneNumber string) (*model.CallCenterCustomer, error) {
	path := "/api/customers/" + url.PathEscape(phoneNumber)
	q := url.Values{}
	h := http.Header{}
	var out *model.CallCenterCustomer
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Statistics(ctx context.Context, days *int) (*model.CallCenterCallStatistics, error) {
	path := "/api/analytics/statistics"
	q := url.Values{}
	if days != nil {
		q.Set("days", strconv.Itoa(*days))
	}
	h := http.Header{}
	var out *model.CallCenterCallStatistics
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
