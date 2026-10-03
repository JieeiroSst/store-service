package payment_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "payment_service"

const DefaultBaseURL = "http://payment-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Payments(ctx context.Context, provider *string, status *string, from *string, to *string, limit *int, offset *int) (*model.PaymentPaymentList, error) {
	path := "/api/v1/payments"
	q := url.Values{}
	if provider != nil {
		q.Set("provider", *provider)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.PaymentPaymentList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Payment(ctx context.Context, id int) (*model.PaymentPayment, error) {
	path := "/api/v1/payments/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.PaymentPayment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transactions(ctx context.Context, id int) ([]*model.PaymentTransaction, error) {
	path := "/api/v1/payments/" + url.PathEscape(strconv.Itoa(id)) + "/transactions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.PaymentTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
