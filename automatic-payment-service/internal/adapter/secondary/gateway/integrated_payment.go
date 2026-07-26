package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/automatic-payment-service/config"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	cbhttp "github.com/JIeeiroSst/utils/circuit_breaker/http"
)

type IntegratedPaymentGateway struct {
	baseURL string
	client  cbhttp.ClientCircuitBreakerProxy
}

func NewIntegratedPaymentGateway(cfg *config.Config) port.PaymentGatewayPort {
	return &IntegratedPaymentGateway{
		baseURL: strings.TrimSuffix(cfg.Gateway.IntegratedPaymentServiceURL, "/"),
		client:  cbhttp.NewClientCircuitBreakerProxy(),
	}
}

type createPaymentRequest struct {
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Description   string  `json:"description"`
}

type paymentResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
	Message       string `json:"message"`
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    paymentResponse `json:"data"`
	Error   string          `json:"error"`
}

func (g *IntegratedPaymentGateway) Charge(ctx context.Context, req port.ChargeRequest) (port.ChargeResult, error) {
	if req.PaymentMethod == nil {
		return port.ChargeResult{Success: false, ErrorMessage: "no payment method provided"}, nil
	}

	body, err := json.Marshal(createPaymentRequest{
		UserID:        req.PaymentMethod.UserID.String(),
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: mapProvider(req.PaymentMethod.Provider),
		Description:   req.Description,
	})
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("gateway: encode request: %w", err)
	}

	raw, err := g.client.Post(ctx, g.baseURL+"/api/v1/payments/", http.MethodPost, body)
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("gateway: call integrated-payment-service: %w", err)
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return port.ChargeResult{}, fmt.Errorf("gateway: decode response: %w", err)
	}

	if !envelope.Success || envelope.Data.Status != "success" {
		msg := envelope.Data.Message
		if msg == "" {
			msg = envelope.Message
		}
		if msg == "" {
			msg = fmt.Sprintf("payment not completed (status=%s)", envelope.Data.Status)
		}
		return port.ChargeResult{Success: false, ErrorMessage: msg}, nil
	}

	return port.ChargeResult{
		Success:              true,
		GatewayTransactionID: envelope.Data.TransactionID,
	}, nil
}

func mapProvider(provider string) string {
	switch strings.ToLower(provider) {
	case "visa", "mastercard", "amex", "card", "stripe":
		return "stripe"
	case "paypal":
		return "paypal"
	case "momo":
		return "momo"
	case "vnpay":
		return "vnpay"
	case "zalopay":
		return "zalopay"
	case "bank_transfer", "bank":
		return "bank_transfer"
	default:
		return provider
	}
}
