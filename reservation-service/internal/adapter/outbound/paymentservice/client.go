package paymentservice

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/JIeeiroSSt/reservation-service/config"
	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/httpx"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type Client struct{ c *httpx.Client }

var _ outbound.PaymentGateway = (*Client)(nil)

func NewClient(cfg config.Config) *Client {
	return &Client{c: httpx.New("payment-service", cfg.PaymentServiceURL, cfg.PaymentServiceTimeout)}
}

type paymentJSON struct {
	ID       int64  `json:"ID"`
	Provider string `json:"Provider"`
	Amount   int64  `json:"Amount"`
	Currency string `json:"Currency"`
	Status   string `json:"Status"`
}

func (p paymentJSON) toDomain() domain.GatewayPayment {
	return domain.GatewayPayment{ID: p.ID, Provider: p.Provider, Amount: p.Amount, Currency: p.Currency, Status: domain.GatewayStatus(p.Status)}
}

func (c *Client) Create(ctx context.Context, in outbound.CreateGatewayPayment) (domain.GatewayPayment, error) {
	var out paymentJSON
	st, msg, err := c.c.Do(ctx, http.MethodPost, "/api/v1/payments", map[string]string{"Idempotency-Key": in.IdempotencyKey}, map[string]any{
		"provider": in.Provider, "amount": in.Amount, "currency": in.Currency,
		"payer_email": in.PayerEmail, "description": in.Description, "idempotency_key": in.IdempotencyKey,
	}, &out)
	if err != nil {
		return domain.GatewayPayment{}, err
	}
	if !httpx.OK(st) {
		return domain.GatewayPayment{}, mapStatus(st, msg)
	}
	return out.toDomain(), nil
}

func (c *Client) Get(ctx context.Context, id int64) (domain.GatewayPayment, error) {
	var out paymentJSON
	st, msg, err := c.c.Do(ctx, http.MethodGet, "/api/v1/payments/"+strconv.FormatInt(id, 10), nil, nil, &out)
	if err != nil {
		return domain.GatewayPayment{}, err
	}
	if !httpx.OK(st) {
		return domain.GatewayPayment{}, mapStatus(st, msg)
	}
	return out.toDomain(), nil
}

func (c *Client) Refund(ctx context.Context, id int64) error {
	p, err := c.Get(ctx, id)
	if err != nil {
		return err
	}
	if p.Status == domain.GatewayRefunded {
		return nil
	}
	st, msg, err := c.c.Do(ctx, http.MethodPost, "/api/v1/payments/"+strconv.FormatInt(id, 10)+"/refund", nil, map[string]int64{"amount": 0}, nil)
	if err != nil {
		return err
	}
	if !httpx.OK(st) {
		return mapStatus(st, msg)
	}
	return nil
}

func (c *Client) RefundPartial(ctx context.Context, id, amount int64) error {
	p, err := c.Get(ctx, id)
	if err != nil {
		return err
	}
	if p.Status == domain.GatewayRefunded || p.Status == domain.GatewayPartially {
		return nil
	}
	st, msg, err := c.c.Do(ctx, http.MethodPost, "/api/v1/payments/"+strconv.FormatInt(id, 10)+"/refund", nil, map[string]int64{"amount": amount}, nil)
	if err != nil {
		return err
	}
	if !httpx.OK(st) {
		return mapStatus(st, msg)
	}
	return nil
}

func mapStatus(status int, msg string) error {
	switch status {
	case http.StatusNotFound:
		return domain.ErrNotFound
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", domain.ErrInvalid, msg)
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", domain.ErrConflict, msg)
	case http.StatusBadGateway:
		return fmt.Errorf("%w: %s", domain.ErrPaymentFailed, msg)
	default:
		return fmt.Errorf("%w: payment-service returned %d: %s", domain.ErrUpstreamUnavailable, status, msg)
	}
}
