package solana_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "solana-service"

const DefaultBaseURL = "http://solana-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Account(ctx context.Context, address string) (*model.SolAccountInfo, error) {
	path := "/api/v1/solana/accounts/" + url.PathEscape(address)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolAccountInfo
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Balance(ctx context.Context, address string) (*model.SolBalance, error) {
	path := "/api/v1/solana/accounts/" + url.PathEscape(address) + "/balance"
	q := url.Values{}
	h := http.Header{}
	var out *model.SolBalance
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transaction(ctx context.Context, signature string) (*model.SolTransaction, error) {
	path := "/api/v1/solana/transactions/" + url.PathEscape(signature)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Program(ctx context.Context, id string) (*model.SolProgram, error) {
	path := "/api/v1/solana/programs/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolProgram
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CircleWalletSet(ctx context.Context, id string) (*model.SolCircleWalletSet, error) {
	path := "/api/v1/circle/wallet-sets/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolCircleWalletSet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CircleWalletSets(ctx context.Context) ([]*model.SolCircleWalletSet, error) {
	path := "/api/v1/circle/wallet-sets"
	q := url.Values{}
	h := http.Header{}
	var out []*model.SolCircleWalletSet
	err := c.rest.Get(ctx, path, q, h, "walletSets", &out)
	return out, err
}

func (c *Client) CircleWallet(ctx context.Context, id string) (*model.SolCircleWallet, error) {
	path := "/api/v1/circle/wallets/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolCircleWallet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CircleWallets(ctx context.Context, walletSetID *string, blockchain *string) ([]*model.SolCircleWallet, error) {
	path := "/api/v1/circle/wallets"
	q := url.Values{}
	if walletSetID != nil {
		q.Set("walletSetId", *walletSetID)
	}
	if blockchain != nil {
		q.Set("blockchain", *blockchain)
	}
	h := http.Header{}
	var out []*model.SolCircleWallet
	err := c.rest.Get(ctx, path, q, h, "wallets", &out)
	return out, err
}

func (c *Client) CircleWalletBalance(ctx context.Context, id string) ([]*model.SolCircleBalance, error) {
	path := "/api/v1/circle/wallets/" + url.PathEscape(id) + "/balance"
	q := url.Values{}
	h := http.Header{}
	var out []*model.SolCircleBalance
	err := c.rest.Get(ctx, path, q, h, "balances", &out)
	return out, err
}

func (c *Client) CircleWalletNFTs(ctx context.Context, id string) ([]*model.SolNFTBalance, error) {
	path := "/api/v1/circle/wallets/" + url.PathEscape(id) + "/nfts"
	q := url.Values{}
	h := http.Header{}
	var out []*model.SolNFTBalance
	err := c.rest.Get(ctx, path, q, h, "nfts", &out)
	return out, err
}

func (c *Client) CircleTransaction(ctx context.Context, id string) (*model.SolCircleTransaction, error) {
	path := "/api/v1/circle/transactions/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolCircleTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CircleEstimateFee(ctx context.Context, walletID string, destinationAddress string, tokenID string, amount string) (*model.SolFeeEstimate, error) {
	path := "/api/v1/circle/transactions/estimate-fee"
	q := url.Values{}
	q.Set("walletId", walletID)
	q.Set("destinationAddress", destinationAddress)
	q.Set("tokenId", tokenID)
	q.Set("amount", amount)
	h := http.Header{}
	var out *model.SolFeeEstimate
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CircleToken(ctx context.Context, id string) (*model.SolToken, error) {
	path := "/api/v1/circle/tokens/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolToken
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CircleTokens(ctx context.Context, blockchain *string) ([]*model.SolToken, error) {
	path := "/api/v1/circle/tokens"
	q := url.Values{}
	if blockchain != nil {
		q.Set("blockchain", *blockchain)
	}
	h := http.Header{}
	var out []*model.SolToken
	err := c.rest.Get(ctx, path, q, h, "tokens", &out)
	return out, err
}

func (c *Client) BridgeTransfer(ctx context.Context, id string) (*model.SolBridgeTransfer, error) {
	path := "/api/v1/bridge/transfers/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.SolBridgeTransfer
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
