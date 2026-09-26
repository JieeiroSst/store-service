package recompenseservice

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

var _ outbound.LoyaltyGateway = (*Client)(nil)

func NewClient(cfg config.Config) *Client {
	return &Client{c: httpx.New("recompense-service", cfg.RecompenseServiceURL, cfg.PaymentServiceTimeout)}
}

func member(id int64) string { return "/api/members/" + strconv.FormatInt(id, 10) + "/points" }

func (c *Client) Earn(ctx context.Context, memberID int64, key string, amount int64) error {
	st, msg, err := c.c.Do(ctx, http.MethodPost, member(memberID)+"/earn", nil,
		map[string]any{"amount": amount, "idempotencyKey": key}, nil)
	if err != nil {
		return err
	}
	if !httpx.OK(st) {
		return mapStatus(st, msg)
	}
	return nil
}

func (c *Client) Status(ctx context.Context, memberID int64) (outbound.Member, error) {
	var out struct {
		Points   int64  `json:"Points"`
		Tier     int    `json:"Tier"`
		TierName string `json:"TierName"`
	}
	st, msg, err := c.c.Do(ctx, http.MethodGet, member(memberID), nil, nil, &out)
	if err != nil {
		return outbound.Member{}, err
	}
	if !httpx.OK(st) {
		return outbound.Member{}, mapStatus(st, msg)
	}
	return outbound.Member{Points: out.Points, Tier: out.Tier, TierName: out.TierName}, nil
}

func mapStatus(status int, msg string) error {
	if status == http.StatusBadRequest {
		return fmt.Errorf("%w: %s", domain.ErrInvalid, msg)
	}
	return fmt.Errorf("%w: recompense-service returned %d: %s", domain.ErrUpstreamUnavailable, status, msg)
}
