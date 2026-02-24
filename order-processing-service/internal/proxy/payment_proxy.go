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

type paymentProxy struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewPaymentProxy(baseURL, apiKey string) PaymentProxy {
	return &paymentProxy{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (p *paymentProxy) ChargePayment(ctx context.Context, req entity.PaymentRequest) (*entity.PaymentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, apperrors.NewProxyError("payment", "ChargePayment", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/payments/charge", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.NewProxyError("payment", "ChargePayment", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.NewProxyError("payment", "ChargePayment", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, apperrors.NewProxyError("payment", "ChargePayment",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result entity.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, apperrors.NewProxyError("payment", "ChargePayment", err)
	}
	return &result, nil
}

func (p *paymentProxy) RefundPayment(ctx context.Context, paymentID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v1/payments/%s/refund", p.baseURL, paymentID), nil)
	if err != nil {
		return apperrors.NewProxyError("payment", "RefundPayment", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return apperrors.NewProxyError("payment", "RefundPayment", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return apperrors.NewProxyError("payment", "RefundPayment",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}
	return nil
}
