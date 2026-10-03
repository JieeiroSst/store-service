package shopify_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "shopify-service"

const DefaultBaseURL = "http://shopify-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Products(ctx context.Context, limit *int, offset *int) (*model.ShopifyProductList, error) {
	path := "/api/v1/products"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.ShopifyProductList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Product(ctx context.Context, id int) (*model.ShopifyProduct, error) {
	path := "/api/v1/products/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ShopifyProduct
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Orders(ctx context.Context, limit *int, offset *int) (*model.ShopifyOrderList, error) {
	path := "/api/v1/orders"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.ShopifyOrderList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Order(ctx context.Context, id int) (*model.ShopifyOrder, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ShopifyOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
