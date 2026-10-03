package voucher_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "voucher-service"

const DefaultBaseURL = "http://voucher-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Voucher(ctx context.Context, id string) (*model.VoucherVoucher, error) {
	path := "/api/v1/vouchers/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VoucherVoucher
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Validate(ctx context.Context, id string, pin *string) (*model.VoucherValidationResult, error) {
	path := "/api/v1/vouchers/" + url.PathEscape(id) + "/validate"
	q := url.Values{}
	if pin != nil {
		q.Set("pin", *pin)
	}
	h := http.Header{}
	var out *model.VoucherValidationResult
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Vouchers(ctx context.Context, ownerID *string) ([]*model.VoucherVoucher, error) {
	path := "/api/v1/vouchers"
	q := url.Values{}
	if ownerID != nil {
		q.Set("owner_id", *ownerID)
	}
	h := http.Header{}
	var out []*model.VoucherVoucher
	err := c.rest.Get(ctx, path, q, h, "vouchers", &out)
	return out, err
}

func (c *Client) Order(ctx context.Context, id string) (*model.VoucherOrder, error) {
	path := "/api/v1/orders/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VoucherOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Merchant(ctx context.Context, id string) (*model.VoucherMerchant, error) {
	path := "/api/v1/merchants/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VoucherMerchant
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Merchants(ctx context.Context) ([]*model.VoucherMerchant, error) {
	path := "/api/v1/merchants"
	q := url.Values{}
	h := http.Header{}
	var out []*model.VoucherMerchant
	err := c.rest.Get(ctx, path, q, h, "merchants", &out)
	return out, err
}

func (c *Client) WalletBalance(ctx context.Context, ownerType string, ownerID string) (*model.VoucherBalance, error) {
	path := "/api/v1/wallets/" + url.PathEscape(ownerType) + "/" + url.PathEscape(ownerID) + "/balance"
	q := url.Values{}
	h := http.Header{}
	var out *model.VoucherBalance
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RedemptionReports(ctx context.Context, from *string, to *string) ([]*model.VoucherRedemptionReport, error) {
	path := "/api/v1/reports/redemptions"
	q := url.Values{}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out []*model.VoucherRedemptionReport
	err := c.rest.Get(ctx, path, q, h, "reports", &out)
	return out, err
}

func (c *Client) CorporateSpend(ctx context.Context, id string, from *string, to *string) (*model.VoucherCorporateSpendReport, error) {
	path := "/api/v1/reports/corporates/" + url.PathEscape(id) + "/spend"
	q := url.Values{}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out *model.VoucherCorporateSpendReport
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
