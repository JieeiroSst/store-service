package integrated_payment_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "integrated-payment-service"

const DefaultBaseURL = "http://integrated-payment-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Payment(ctx context.Context, id string) (*model.IntPayPayment, error) {
	path := "/api/v1/payments/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.IntPayPayment
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) PaymentStatus(ctx context.Context, id string) (*string, error) {
	path := "/api/v1/payments/" + url.PathEscape(id) + "/status"
	q := url.Values{}
	h := http.Header{}
	var out *string
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
