package paymentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/doordash-service/config"
	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type httpPaymentClient struct {
	client  *http.Client
	baseURL string
}

func NewPaymentClient(cfg *config.Config) port.PaymentClient {
	return &httpPaymentClient{
		client:  &http.Client{Timeout: cfg.PaymentService.TimeoutDuration()},
		baseURL: cfg.PaymentService.BaseURL,
	}
}

type authorizePaymentRequest struct {
	CustomerID      string  `json:"customer_id"`
	Amount          float64 `json:"amount"`
	PaymentMethodID string  `json:"payment_method_id"`
}

type authorizePaymentResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

// AuthorizePayment places a hold on the customer's payment method for the
// order total. A non-2xx response, and any status the payment service
// returns other than "authorized", both surface as
// port.ErrPaymentAuthorizationFailed via application.orderService - this
// method itself returns the raw response so the caller can log the actual
// status.
func (c *httpPaymentClient) AuthorizePayment(ctx context.Context, in port.AuthorizePaymentInput) (*model.PaymentAuthorization, error) {
	payload, err := json.Marshal(authorizePaymentRequest{
		CustomerID:      in.CustomerID,
		Amount:          in.Amount,
		PaymentMethodID: in.PaymentMethodID,
	})
	if err != nil {
		return nil, fmt.Errorf("encode payment authorization request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/payments/authorize", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call payment service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return &model.PaymentAuthorization{Status: model.PaymentStatusFailed}, nil
	}

	var body authorizePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode payment service response: %w", err)
	}
	return &model.PaymentAuthorization{PaymentID: body.PaymentID, Status: body.Status}, nil
}
