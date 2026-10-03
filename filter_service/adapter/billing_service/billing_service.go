package billing_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "billing-service"

const DefaultBaseURL = "http://billing-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Addresses(ctx context.Context) ([]*model.BillingAddress, error) {
	path := "/addresses"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BillingAddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Address(ctx context.Context, id int) (*model.BillingAddress, error) {
	path := "/addresses/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BillingAddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Customers(ctx context.Context) ([]*model.BillingCustomer, error) {
	path := "/customers"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BillingCustomer
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Customer(ctx context.Context, id int) (*model.BillingCustomer, error) {
	path := "/customers/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BillingCustomer
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Plans(ctx context.Context) ([]*model.BillingPlan, error) {
	path := "/plans"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BillingPlan
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Plan(ctx context.Context, id int) (*model.BillingPlan, error) {
	path := "/plans/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BillingPlan
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Subscriptions(ctx context.Context) ([]*model.BillingSubscription, error) {
	path := "/subscriptions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BillingSubscription
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Subscription(ctx context.Context, id int) (*model.BillingSubscription, error) {
	path := "/subscriptions/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BillingSubscription
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Invoices(ctx context.Context) ([]*model.BillingInvoice, error) {
	path := "/invoices"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BillingInvoice
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Invoice(ctx context.Context, id int) (*model.BillingInvoice, error) {
	path := "/invoices/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BillingInvoice
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Transactions(ctx context.Context) ([]*model.BillingTransaction, error) {
	path := "/transactions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BillingTransaction
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Transaction(ctx context.Context, id int) (*model.BillingTransaction, error) {
	path := "/transactions/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BillingTransaction
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
