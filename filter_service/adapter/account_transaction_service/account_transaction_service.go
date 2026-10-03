package account_transaction_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "account-transaction-service"

const DefaultBaseURL = "http://account-transaction-api-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) GetAccount(ctx context.Context, id string) (*model.AcctTxAccount, error) {
	path := "/v1/accounts/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AcctTxAccount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListAccounts(ctx context.Context, pageSize *int, pageToken *string) (*model.AcctTxListAccountsResponse, error) {
	path := "/v1/accounts"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	h := http.Header{}
	var out *model.AcctTxListAccountsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetTransaction(ctx context.Context, id string) (*model.AcctTxTransaction, error) {
	path := "/v1/transactions/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AcctTxTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListTransactions(ctx context.Context, pageSize *int, pageToken *string, typeArg *string) (*model.AcctTxListTransactionsResponse, error) {
	path := "/v1/transactions"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	h := http.Header{}
	var out *model.AcctTxListTransactionsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetAccountTransactions(ctx context.Context, accountID string, pageSize *int, pageToken *string) (*model.AcctTxListTransactionsResponse, error) {
	path := "/v1/accounts/" + url.PathEscape(accountID) + "/transactions"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	h := http.Header{}
	var out *model.AcctTxListTransactionsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
