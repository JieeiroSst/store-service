package restaurant_consumer_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "restaurant-consumer-service"

const DefaultBaseURL = "http://consumer-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Consumers(ctx context.Context, limit *int, page *int, sort *string) (*model.RConsumerPage, error) {
	path := "/api/v1/consumer"
	q := url.Values{}
	if limit != nil {
		q.Set("Limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("Page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("Sort", *sort)
	}
	h := http.Header{}
	var out *model.RConsumerPage
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
