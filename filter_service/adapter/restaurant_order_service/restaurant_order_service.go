package restaurant_order_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "restaurant-order-service"

const DefaultBaseURL = "http://order-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Order(ctx context.Context, id int) (*model.ROrderOrder, error) {
	path := "/api/v1/order/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ROrderOrder
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Orders(ctx context.Context, limit *int, page *int, sort *string) (*model.ROrderOrderPage, error) {
	path := "/api/v1/order"
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
	var out *model.ROrderOrderPage
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Reservations(ctx context.Context, limit *int, page *int, sort *string) (*model.ROrderReservationPage, error) {
	path := "/api/v1/reservations"
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
	var out *model.ROrderReservationPage
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Reservation(ctx context.Context, id int) (*model.ROrderReservation, error) {
	path := "/api/v1/reservations/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ROrderReservation
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
