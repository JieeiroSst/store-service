package shipping

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/config"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"go.uber.org/fx"
)

const (
	ItemWeightGram  = 500
	maxStackedUnits = 6
	maxInsurance    = 5_000_000
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, http: &http.Client{Timeout: timeout}}
}

var Module = fx.Options(fx.Provide(func(cfg *config.Config) port.ShippingRates {
	s := cfg.ShippingService
	if s.BaseURL == "" || s.APIKey == "" {
		return nil
	}
	return New(s.BaseURL, s.APIKey, time.Duration(s.TimeoutSeconds)*time.Second)
}))

type item struct {
	Name       string `json:"name"`
	Code       string `json:"code,omitempty"`
	Quantity   int    `json:"quantity"`
	WeightGram int    `json:"weight_gram"`
	Price      int64  `json:"price"`
}

type quoteRequest struct {
	Recipient struct {
		Name       string `json:"name"`
		Phone      string `json:"phone"`
		Street     string `json:"street"`
		WardCode   string `json:"ward_code"`
		DistrictID int    `json:"district_id"`
		ProvinceID int    `json:"province_id"`
	} `json:"recipient"`
	Parcel struct {
		WeightGram int    `json:"weight_gram"`
		LengthCm   int    `json:"length_cm"`
		WidthCm    int    `json:"width_cm"`
		HeightCm   int    `json:"height_cm"`
		Items      []item `json:"items"`
	} `json:"parcel"`
	InsuranceValue int64 `json:"insurance_value"`
}

type option struct {
	WarehouseID        int64  `json:"warehouse_id"`
	ServiceID          int    `json:"service_id"`
	Fee                int64  `json:"fee"`
	ServiceName        string `json:"service_name"`
	ExpectedDeliveryAt string `json:"expected_delivery_at"`
}

func Request(to port.ShippingAddress, items []port.ShippingItem, insurance int64) quoteRequest {
	var r quoteRequest
	r.Recipient.Name, r.Recipient.Phone, r.Recipient.Street = to.Name, to.Phone, to.Street
	r.Recipient.WardCode, r.Recipient.DistrictID, r.Recipient.ProvinceID = to.WardCode, to.DistrictID, to.ProvinceID
	units := 0
	for _, it := range items {
		r.Parcel.Items = append(r.Parcel.Items, item{Name: it.Name, Code: it.Code, Quantity: it.Quantity, WeightGram: ItemWeightGram, Price: it.Price})
		units += it.Quantity
	}
	units = max(units, 1)
	r.Parcel.WeightGram = units * ItemWeightGram
	r.Parcel.LengthCm, r.Parcel.WidthCm = 30, 20
	r.Parcel.HeightCm = 10 + 5*min(units, maxStackedUnits)
	r.InsuranceValue = min(max(insurance, 0), maxInsurance)
	return r
}

func (c *Client) Quote(ctx context.Context, to port.ShippingAddress, items []port.ShippingItem, insurance int64) (*port.ShippingQuote, error) {
	body, err := json.Marshal(Request(to, items, insurance))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/shipments/quote", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: shipping service: %v", common.ErrUpstream, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	switch {
	case resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity:
		var failure struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &failure)
		return nil, fmt.Errorf("%w: không giao được tới địa chỉ này (%s)", common.ErrInvalidRequest, failure.Error)
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: shipping service returned %d", common.ErrUpstream, resp.StatusCode)
	}
	var out struct {
		Options     []option           `json:"options"`
		Recommended map[string]*option `json:"recommended"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%w: shipping service: %v", common.ErrUpstream, err)
	}
	best := out.Recommended["CHEAPEST"]
	for i := range out.Options {
		if best == nil || out.Options[i].Fee < best.Fee {
			best = &out.Options[i]
		}
	}
	if best == nil {
		return nil, fmt.Errorf("%w: không có tuyến giao hàng tới địa chỉ này", common.ErrInvalidRequest)
	}
	return &port.ShippingQuote{Fee: best.Fee, WarehouseID: best.WarehouseID, ServiceID: best.ServiceID, ServiceName: best.ServiceName, ExpectedDeliveryAt: best.ExpectedDeliveryAt}, nil
}
