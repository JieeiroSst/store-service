package sard_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "sard-service"

const DefaultBaseURL = "http://sard-customer-info-service:8081"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Info(ctx context.Context) (*model.SardInfo, error) {
	path := "/api/info"
	q := url.Values{}
	h := http.Header{}
	var out *model.SardInfo
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
