package booking_mini_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "booking-mini-service"

const DefaultBaseURL = "http://booking-mini-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) AvailableRooms(ctx context.Context, fromDate string, toDate string) ([]*model.BookingMiniRoom, error) {
	path := "/api/rooms"
	q := url.Values{}
	q.Set("from_date", fromDate)
	q.Set("to_date", toDate)
	h := http.Header{}
	var out []*model.BookingMiniRoom
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Booking(ctx context.Context, id int) (*model.BookingMiniBooking, error) {
	path := "/api/bookings/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookingMiniBooking
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Passenger(ctx context.Context, id int) (*model.BookingMiniPassenger, error) {
	path := "/api/passengers/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BookingMiniPassenger
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
