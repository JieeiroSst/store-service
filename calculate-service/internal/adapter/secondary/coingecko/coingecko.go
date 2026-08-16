package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
)

type client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) port.MarketProvider {
	return &client{baseURL: baseURL, httpClient: httpClient}
}

func (c *client) GetMarkets(ctx context.Context, vsCurrency string, perPage int) ([]model.Coin, error) {
	if perPage <= 0 {
		perPage = 100
	}

	q := url.Values{}
	q.Set("vs_currency", vsCurrency)
	q.Set("per_page", strconv.Itoa(perPage))
	q.Set("page", "1")

	reqURL := c.baseURL + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko: unexpected status %d", resp.StatusCode)
	}

	var coins []model.Coin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, err
	}

	return coins, nil
}
