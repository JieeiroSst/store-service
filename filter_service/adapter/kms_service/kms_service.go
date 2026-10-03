package kms_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "kms-service"

const DefaultBaseURL = "http://kms-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Keys(ctx context.Context) ([]*model.KmsKey, error) {
	path := "/api/v1/keys"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KmsKey
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Key(ctx context.Context, id string) (*model.KmsKey, error) {
	path := "/api/v1/keys/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.KmsKey
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) KeyForUse(ctx context.Context, id string) (*model.KmsKeyUse, error) {
	path := "/api/v1/keys/" + url.PathEscape(id) + "/use"
	q := url.Values{}
	h := http.Header{}
	var out *model.KmsKeyUse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) KeyUsageStats(ctx context.Context, id string) (*model.KmsKeyStats, error) {
	path := "/api/v1/keys/" + url.PathEscape(id) + "/stats"
	q := url.Values{}
	h := http.Header{}
	var out *model.KmsKeyStats
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) KeyAuditLogs(ctx context.Context, id string, limit *int, offset *int) (*model.KmsKeyAuditLogs, error) {
	path := "/api/v1/keys/" + url.PathEscape(id) + "/audit"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.KmsKeyAuditLogs
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuditLogs(ctx context.Context, limit *int, offset *int) (*model.KmsAuditLogs, error) {
	path := "/api/v1/audit/logs"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.KmsAuditLogs
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
