package restaurant_delivery_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "restaurant-delivery-service"

const DefaultBaseURL = "http://delivery-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Deliveries(ctx context.Context, limit *int, page *int, sort *string) (*model.RDeliveryPage, error) {
	path := "/api/v1/delivery/"
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
	var out *model.RDeliveryPage
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
