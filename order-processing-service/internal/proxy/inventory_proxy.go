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

type inventoryProxy struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewInventoryProxy(baseURL, apiKey string) InventoryProxy {
	return &inventoryProxy{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (p *inventoryProxy) ReserveStock(ctx context.Context, req entity.InventoryReserveRequest) (*entity.InventoryReserveResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, apperrors.NewProxyError("inventory", "ReserveStock", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/inventory/reserve", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.NewProxyError("inventory", "ReserveStock", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.NewProxyError("inventory", "ReserveStock", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, apperrors.NewProxyError("inventory", "ReserveStock",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result entity.InventoryReserveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, apperrors.NewProxyError("inventory", "ReserveStock", err)
	}
	return &result, nil
}

func (p *inventoryProxy) ReleaseStock(ctx context.Context, orderID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v1/inventory/release/%s", p.baseURL, orderID), nil)
	if err != nil {
		return apperrors.NewProxyError("inventory", "ReleaseStock", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return apperrors.NewProxyError("inventory", "ReleaseStock", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return apperrors.NewProxyError("inventory", "ReleaseStock",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}
	return nil
}
