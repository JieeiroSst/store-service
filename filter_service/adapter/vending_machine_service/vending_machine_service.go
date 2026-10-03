package vending_machine_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "vending-machine-service"

const DefaultBaseURL = "http://vending-machine-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Machines(ctx context.Context, status *string, cursor *string, limit *int) (*model.VendMachinePage, error) {
	path := "/api/v1/machines"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.VendMachinePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Machine(ctx context.Context, id string) (*model.VendMachine, error) {
	path := "/api/v1/machines/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VendMachine
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Maintenance(ctx context.Context, id string) ([]*model.VendMaintenance, error) {
	path := "/api/v1/machines/" + url.PathEscape(id) + "/maintenance"
	q := url.Values{}
	h := http.Header{}
	var out []*model.VendMaintenance
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Events(ctx context.Context, id string, limit *int) ([]*model.VendEvent, error) {
	path := "/api/v1/machines/" + url.PathEscape(id) + "/events"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.VendEvent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SalesReport(ctx context.Context, id string, from *string, to *string) (*model.VendSalesReport, error) {
	path := "/api/v1/machines/" + url.PathEscape(id) + "/sales"
	q := url.Values{}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out *model.VendSalesReport
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Inventory(ctx context.Context, id string, cursor *string, limit *int) (*model.VendInventoryPage, error) {
	path := "/api/v1/machines/" + url.PathEscape(id) + "/inventory"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.VendInventoryPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) LowInventory(ctx context.Context, machineID *string, cursor *string, limit *int) (*model.VendInventoryPage, error) {
	path := "/api/v1/inventory/low"
	q := url.Values{}
	if machineID != nil {
		q.Set("machine_id", *machineID)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.VendInventoryPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Categories(ctx context.Context, cursor *string, limit *int) (*model.VendCategoryPage, error) {
	path := "/api/v1/categories"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.VendCategoryPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Products(ctx context.Context, categoryID *string, cursor *string, limit *int) (*model.VendProductPage, error) {
	path := "/api/v1/products"
	q := url.Values{}
	if categoryID != nil {
		q.Set("category_id", *categoryID)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.VendProductPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Product(ctx context.Context, id string) (*model.VendProduct, error) {
	path := "/api/v1/products/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VendProduct
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Session(ctx context.Context, id string) (*model.VendSession, error) {
	path := "/api/v1/sessions/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VendSession
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SessionOrders(ctx context.Context, id string) ([]*model.VendOrder, error) {
	path := "/api/v1/sessions/" + url.PathEscape(id) + "/orders"
	q := url.Values{}
	h := http.Header{}
	var out []*model.VendOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Payment(ctx context.Context, id string) (*model.VendPaymentStatus, error) {
	path := "/api/v1/payments/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VendPaymentStatus
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Order(ctx context.Context, id string) (*model.VendOrder, error) {
	path := "/api/v1/orders/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VendOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
