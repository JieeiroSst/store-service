package authorize_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "authorize_service"

const DefaultBaseURL = "http://authorize-service-api-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) GetCasbinRules(ctx context.Context, limit *int, page *int, sort *string, totalRows *string, totalPages *int) (*model.AuthzCasbinRuleList, error) {
	path := "/casbin"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if totalRows != nil {
		q.Set("total_rows", *totalRows)
	}
	if totalPages != nil {
		q.Set("total_pages", strconv.Itoa(*totalPages))
	}
	h := http.Header{}
	var out *model.AuthzCasbinRuleList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetCasbinRuleByID(ctx context.Context, id int) (*model.AuthzCasbinRule, error) {
	path := "/casbin/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AuthzCasbinRule
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
