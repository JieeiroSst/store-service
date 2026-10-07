package customerinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JIeeiroSst/card-service/config"
	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/JIeeiroSst/card-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) port.CustomerDirectory {
		if cfg.CustomerInfo.BaseURL == "" {
			log.Printf("customer-info-service is not configured: every customer is treated as eligible (development only)")
			return Unchecked{}
		}
		return NewClient(cfg.CustomerInfo.BaseURL, cfg.CustomerInfo.Token, cfg.CustomerInfo.Timeout)
	}),
)

const maxResponseBytes = 1 << 20

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout}}
}

type customerResponse struct {
	ID              string `json:"id"`
	UserID          int64  `json:"user_id"`
	FullName        string `json:"full_name"`
	CardEligibility struct {
		Eligible bool     `json:"eligible"`
		Reasons  []string `json:"reasons"`
	} `json:"card_eligibility"`
}

func (c *Client) GetCustomer(ctx context.Context, id string) (domain.Customer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/customers/"+url.PathEscape(id), nil)
	if err != nil {
		return domain.Customer{}, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return domain.Customer{}, fmt.Errorf("%w: customer-info-service: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return domain.Customer{}, fmt.Errorf("%w: customer-info-service: %v", domain.ErrUnavailable, err)
	}
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return domain.Customer{}, domain.Invalid("customer %s does not exist in customer-info-service", id)
	case resp.StatusCode != http.StatusOK:
		return domain.Customer{}, fmt.Errorf("%w: customer-info-service returned %s", domain.ErrUnavailable, resp.Status)
	}
	var cr customerResponse
	if err := json.Unmarshal(raw, &cr); err != nil || cr.ID != id {
		return domain.Customer{}, fmt.Errorf("%w: customer-info-service: bad customer response", domain.ErrUnavailable)
	}
	return domain.Customer{
		ID: cr.ID, UserID: cr.UserID, FullName: cr.FullName,
		Eligible: cr.CardEligibility.Eligible, Reasons: cr.CardEligibility.Reasons,
	}, nil
}

type Unchecked struct{}

func (Unchecked) GetCustomer(_ context.Context, id string) (domain.Customer, error) {
	return domain.Customer{ID: id, FullName: "SARD CARDHOLDER", Eligible: true}, nil
}
