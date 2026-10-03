package arrange_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "arrange-service"

const DefaultBaseURL = "http://arrange-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Shifts(ctx context.Context, storeID int) ([]*model.ArrangeShift, error) {
	path := "/shifts"
	q := url.Values{}
	q.Set("storeId", strconv.Itoa(storeID))
	h := http.Header{}
	var out []*model.ArrangeShift
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AvailableSpaces(ctx context.Context, storeID int, minCapacity *int) ([]*model.ArrangeSpace, error) {
	path := "/spaces"
	q := url.Values{}
	q.Set("storeId", strconv.Itoa(storeID))
	if minCapacity != nil {
		q.Set("minCapacity", strconv.Itoa(*minCapacity))
	}
	h := http.Header{}
	var out []*model.ArrangeSpace
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SpaceAssignments(ctx context.Context, storeID int) ([]*model.ArrangeSpaceAssignment, error) {
	path := "/spaces/assignments"
	q := url.Values{}
	q.Set("storeId", strconv.Itoa(storeID))
	h := http.Header{}
	var out []*model.ArrangeSpaceAssignment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) DispatchAssignments(ctx context.Context, driverID int) ([]*model.ArrangeDispatchAssignment, error) {
	path := "/dispatch/assignments"
	q := url.Values{}
	q.Set("driverId", strconv.Itoa(driverID))
	h := http.Header{}
	var out []*model.ArrangeDispatchAssignment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
