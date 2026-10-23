package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
)

var (
	ErrShipmentRejected = errors.New("shipment rejected")
	ErrShipmentInvalid  = errors.New("shipment request invalid")
	ErrShipmentAuth     = errors.New("shipping-service refused the API key")
)

type ShipmentProxy interface {
	CreateShipment(ctx context.Context, req entity.ShipmentRequest) (*entity.Shipment, error)
}

type shipmentProxy struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewShipmentProxy(baseURL, apiKey string) ShipmentProxy {
	return &shipmentProxy{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *shipmentProxy) CreateShipment(ctx context.Context, req entity.ShipmentRequest) (*entity.Shipment, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/shipments", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("shipping-service: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	switch {
	case resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated:
		var shipment entity.Shipment
		if err := json.Unmarshal(raw, &shipment); err != nil {
			return nil, fmt.Errorf("shipping-service: decode shipment: %w", err)
		}
		return &shipment, nil
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, ErrShipmentAuth
	case resp.StatusCode == http.StatusUnprocessableEntity:
		var failure struct {
			Error    string           `json:"error"`
			Shipment *entity.Shipment `json:"shipment"`
		}
		_ = json.Unmarshal(raw, &failure)
		return failure.Shipment, fmt.Errorf("%w: %s", ErrShipmentRejected, failure.Error)
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		var failure struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &failure)
		if failure.Error == "" {
			failure.Error = strings.TrimSpace(string(raw))
		}
		return nil, fmt.Errorf("%w: %d %s", ErrShipmentInvalid, resp.StatusCode, failure.Error)
	default:
		return nil, fmt.Errorf("shipping-service: status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
}
