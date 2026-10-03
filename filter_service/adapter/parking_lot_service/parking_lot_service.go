package parking_lot_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "parking-lot-service"

const DefaultBaseURL = "http://parking-lot-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) AvailableSpots(ctx context.Context) ([]*model.ParkingSpotAvailability, error) {
	path := "/api/v1/spots/available"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ParkingSpotAvailability
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Rates(ctx context.Context) ([]*model.ParkingRate, error) {
	path := "/api/v1/rates"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ParkingRate
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Ticket(ctx context.Context, id string) (*model.ParkingTicket, error) {
	path := "/api/v1/tickets/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.ParkingTicket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) History(ctx context.Context, plate string, limit *int, offset *int) ([]*model.ParkingTicket, error) {
	path := "/api/v1/history"
	q := url.Values{}
	q.Set("plate", plate)
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.ParkingTicket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
