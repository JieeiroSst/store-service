package automatic_payment_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "automatic-payment-service"

const DefaultBaseURL = "http://automatic-payment-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Subscription(ctx context.Context, id string) (*model.AutoPaySubscription, error) {
	path := "/subscriptions/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AutoPaySubscription
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) SubscriptionTransactions(ctx context.Context, id string) ([]*model.AutoPayTransaction, error) {
	path := "/subscriptions/" + url.PathEscape(id) + "/transactions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AutoPayTransaction
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) SubscriptionInvoices(ctx context.Context, id string) ([]*model.AutoPayInvoice, error) {
	path := "/subscriptions/" + url.PathEscape(id) + "/invoices"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AutoPayInvoice
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) PaymentMethods(ctx context.Context, userID string) ([]*model.AutoPayPaymentMethod, error) {
	path := "/payment-methods"
	q := url.Values{}
	q.Set("userId", userID)
	h := http.Header{}
	var out []*model.AutoPayPaymentMethod
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Invoices(ctx context.Context, userID string) ([]*model.AutoPayInvoice, error) {
	path := "/invoices"
	q := url.Values{}
	q.Set("userId", userID)
	h := http.Header{}
	var out []*model.AutoPayInvoice
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Invoice(ctx context.Context, id string) (*model.AutoPayInvoice, error) {
	path := "/invoices/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AutoPayInvoice
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
