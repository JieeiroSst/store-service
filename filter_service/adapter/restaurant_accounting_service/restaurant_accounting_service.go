package restaurant_accounting_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "restaurant-accounting-service"

const DefaultBaseURL = "http://accounting-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Payment(ctx context.Context, orderID int) (*model.RAccountingPayment, error) {
	path := "/api/v1/payments/" + url.PathEscape(strconv.Itoa(orderID))
	q := url.Values{}
	h := http.Header{}
	var out *model.RAccountingPayment
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
