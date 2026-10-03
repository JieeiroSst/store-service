package doordash_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "doordash-service"

const DefaultBaseURL = "http://doordash-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Orders(ctx context.Context, customerID *string, restaurantID *string) ([]*model.DoordashOrder, error) {
	path := "/api/v1/orders"
	q := url.Values{}
	if customerID != nil {
		q.Set("customer_id", *customerID)
	}
	if restaurantID != nil {
		q.Set("restaurant_id", *restaurantID)
	}
	h := http.Header{}
	var out []*model.DoordashOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Order(ctx context.Context, id int) (*model.DoordashOrder, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.DoordashOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) OrderTracking(ctx context.Context, id int) ([]*model.DoordashOrderTracking, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(id)) + "/tracking"
	q := url.Values{}
	h := http.Header{}
	var out []*model.DoordashOrderTracking
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) DriverAssignments(ctx context.Context, driverID string) ([]*model.DoordashDriverAssignment, error) {
	path := "/api/v1/driver-assignments"
	q := url.Values{}
	q.Set("driver_id", driverID)
	h := http.Header{}
	var out []*model.DoordashDriverAssignment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
