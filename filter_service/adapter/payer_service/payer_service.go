package payer_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "payer-service"

const DefaultBaseURL = "http://payer-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Transaction(ctx context.Context, transactionID int) (*model.PayerTransactionResponse, error) {
	path := "/api/v1/transaction"
	q := url.Values{}
	q.Set("transaction_id", strconv.Itoa(transactionID))
	h := http.Header{}
	var out *model.PayerTransactionResponse
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
