package bot_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "bot-service"

const DefaultBaseURL = "http://bot-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Posts(ctx context.Context, limit *int, offset *int, status *string, campaign *string) ([]*model.BotPost, error) {
	path := "/bot-content/v1/posts"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if campaign != nil {
		q.Set("campaign", *campaign)
	}
	h := http.Header{}
	var out []*model.BotPost
	err := c.rest.Get(ctx, path, q, h, "posts", &out)
	return out, err
}

func (c *Client) Post(ctx context.Context, id string) (*model.BotPost, error) {
	path := "/bot-content/v1/posts/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.BotPost
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
