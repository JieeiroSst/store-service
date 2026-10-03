package payment_wallet_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "payment-wallet-service"

const DefaultBaseURL = "http://payment-wallet-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Wallet(ctx context.Context, id string) (*model.WalletWallet, error) {
	path := "/api/v1/wallets/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletWallet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) WalletByUser(ctx context.Context, userID string) (*model.WalletWallet, error) {
	path := "/api/v1/wallets/user/" + url.PathEscape(userID)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletWallet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transactions(ctx context.Context, id string, limit *int, offset *int) ([]*model.WalletTransaction, error) {
	path := "/api/v1/wallets/" + url.PathEscape(id) + "/transactions"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.WalletTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) TransactionByReference(ctx context.Context, id string, referenceID string) (*model.WalletTransaction, error) {
	path := "/api/v1/wallets/" + url.PathEscape(id) + "/transactions/by-reference/" + url.PathEscape(referenceID)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Pockets(ctx context.Context, id string) ([]*model.WalletPocket, error) {
	path := "/api/v1/wallets/" + url.PathEscape(id) + "/pockets"
	q := url.Values{}
	h := http.Header{}
	var out []*model.WalletPocket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transfer(ctx context.Context, id string) (*model.WalletTransfer, error) {
	path := "/api/v1/transfers/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletTransfer
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transaction(ctx context.Context, id string) (*model.WalletTransaction, error) {
	path := "/api/v1/transactions/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PaymentRequests(ctx context.Context, walletID string) ([]*model.WalletPaymentRequest, error) {
	path := "/api/v1/payment-requests"
	q := url.Values{}
	q.Set("wallet_id", walletID)
	h := http.Header{}
	var out []*model.WalletPaymentRequest
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PaymentRequest(ctx context.Context, id string) (*model.WalletPaymentRequest, error) {
	path := "/api/v1/payment-requests/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletPaymentRequest
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PaymentRequestQR(ctx context.Context, id string) (*model.WalletPaymentRequestQR, error) {
	path := "/api/v1/payment-requests/" + url.PathEscape(id) + "/qr"
	q := url.Values{}
	h := http.Header{}
	var out *model.WalletPaymentRequestQR
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PaymentMethods(ctx context.Context, userID string) ([]*model.WalletPaymentMethod, error) {
	path := "/api/v1/payment-methods"
	q := url.Values{}
	q.Set("user_id", userID)
	h := http.Header{}
	var out []*model.WalletPaymentMethod
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
