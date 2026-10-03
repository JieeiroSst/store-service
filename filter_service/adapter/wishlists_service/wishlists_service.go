package wishlists_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "wishlists-service"

const DefaultBaseURL = "http://wishlists-service:8000"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Greet(ctx context.Context) (*string, error) {
	path := "/greet"
	q := url.Values{}
	h := http.Header{}
	var out *string
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Customers(ctx context.Context) ([]*model.WishlistCustomer, error) {
	path := "/customer"
	q := url.Values{}
	h := http.Header{}
	var out []*model.WishlistCustomer
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
