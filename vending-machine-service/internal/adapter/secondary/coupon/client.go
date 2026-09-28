package coupon

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

type Client struct {
	client *httpx.Client
}

func NewClient(client *httpx.Client) *Client { return &Client{client: client} }

func toUnits(cents int) float64 { return float64(cents) / 100 }

func toCents(units float64) int { return int(math.Round(units * 100)) }

func (c *Client) Quote(ctx context.Context, code string, amountCents int) (int, error) {
	var resp struct {
		Discount float64 `json:"discount"`
	}
	err := c.client.Do(ctx, http.MethodPost, "/api/v1/coupons/validate", map[string]any{
		"code":            code,
		"purchase_amount": toUnits(amountCents),
	}, &resp)
	if err != nil {
		return 0, mapErr(err)
	}
	return toCents(resp.Discount), nil
}

func (c *Client) Redeem(ctx context.Context, code string, orderNo int64, amountCents int) error {
	err := c.client.Do(ctx, http.MethodPost, "/api/v1/coupons/apply", map[string]any{
		"code":            code,
		"purchase_amount": toUnits(amountCents),
		"order_id":        orderNo,
	}, nil)
	return mapErr(err)
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var se *httpx.StatusError
	if errors.As(err, &se) && se.Status < 500 {
		return fmt.Errorf("%w: %s", domain.ErrCouponRejected, se.Message)
	}
	return fmt.Errorf("%w: coupon service: %v", domain.ErrUpstream, err)
}

type Disabled struct{}

func (Disabled) Quote(context.Context, string, int) (int, error) {
	return 0, fmt.Errorf("%w: coupons are not available", domain.ErrCouponRejected)
}

func (Disabled) Redeem(context.Context, string, int64, int) error { return nil }
