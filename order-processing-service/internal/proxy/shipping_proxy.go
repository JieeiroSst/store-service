package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	apperrors "github.com/JIeeiroSst/order-processing-service/pkg/errors"
)

type shippingProxy struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewShippingProxy(baseURL, apiKey string) ShippingProxy {
	return &shippingProxy{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (p *shippingProxy) CreateShipment(ctx context.Context, req entity.ShippingRequest) (*entity.ShippingResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, apperrors.NewProxyError("shipping", "CreateShipment", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/shipments", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.NewProxyError("shipping", "CreateShipment", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.NewProxyError("shipping", "CreateShipment", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, apperrors.NewProxyError("shipping", "CreateShipment",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result entity.ShippingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, apperrors.NewProxyError("shipping", "CreateShipment", err)
	}
	return &result, nil
}

func (p *shippingProxy) CancelShipment(ctx context.Context, shippingID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/v1/shipments/%s", p.baseURL, shippingID), nil)
	if err != nil {
		return apperrors.NewProxyError("shipping", "CancelShipment", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return apperrors.NewProxyError("shipping", "CancelShipment", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return apperrors.NewProxyError("shipping", "CancelShipment",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}
	return nil
}
