package wallet_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "wallet-service"

const DefaultBaseURL = "http://wallet-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Wallet(ctx context.Context, id string) (*model.WalletSvcWallet, error) {
	path := "/api/v1/wallets/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletSvcWallet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transaction(ctx context.Context, id string) (*model.WalletSvcTransaction, error) {
	path := "/api/v1/transactions/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletSvcTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
