package catalogue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/config"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

var Module = fx.Options(fx.Provide(func(cfg *config.Config) port.Catalogue {
	return New(cfg.CatalogueService.BaseURL, time.Duration(cfg.CatalogueService.TimeoutSeconds)*time.Second)
}))

type stockRecordJSON struct {
	ID            int     `json:"id"`
	ProductID     int     `json:"product_id"`
	PriceCurrency string  `json:"price_currency"`
	Price         float64 `json:"price"`
	NumAvailable  *int    `json:"num_available"`
}

func (c *Client) StockRecord(ctx context.Context, id int) (*port.StockRecord, error) {
	var body stockRecordJSON
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/v1/stockrecords/%d", id), nil, &body); err != nil {
		return nil, err
	}
	return &port.StockRecord{
		ID: body.ID, ProductID: body.ProductID, PriceCurrency: body.PriceCurrency, Price: body.Price, NumAvailable: body.NumAvailable,
	}, nil
}

func (c *Client) Allocate(ctx context.Context, id, quantity int) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/stockrecords/%d/allocate", id), map[string]int{"quantity": quantity}, nil)
}

func (c *Client) Release(ctx context.Context, id, quantity int) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/stockrecords/%d/release", id), map[string]int{"quantity": quantity}, nil)
}

func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body bytes.Buffer
	if in != nil {
		if err := json.NewEncoder(&body).Encode(in); err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, &body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: catalogue: %v", common.ErrUpstream, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("%w: stock record", common.ErrNotFound)
	case resp.StatusCode == http.StatusConflict:
		return common.ErrOutOfStock
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		return fmt.Errorf("%w: catalogue rejected the request (%d)", common.ErrInvalidRequest, resp.StatusCode)
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: catalogue returned %d", common.ErrUpstream, resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
