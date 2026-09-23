package payoneer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/payment-service/config"
	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
)

type gateway struct {
	client        *http.Client
	baseURL       string
	apiKey        string
	programID     string
	webhookSecret string
}

func NewGateway(cfg *config.Config) port.PaymentGateway {
	return &gateway{
		client:        &http.Client{Timeout: cfg.Payoneer.TimeoutDuration()},
		baseURL:       cfg.Payoneer.BaseURL,
		apiKey:        cfg.Payoneer.APIKey,
		programID:     cfg.Payoneer.ProgramID,
		webhookSecret: cfg.Payoneer.WebhookSecret,
	}
}

func (g *gateway) Provider() model.Provider { return model.ProviderPayoneer }

type payoutResponse struct {
	PayoutID string `json:"payout_id"`
	Status   string `json:"status"`
}

func (g *gateway) CreatePayment(ctx context.Context, in port.CreateGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	payload := map[string]any{
		"payee_id":            in.PayerEmail,
		"amount":              float64(in.Amount) / 100,
		"currency":            strings.ToUpper(in.Currency),
		"description":         in.Description,
		"client_reference_id": in.ReferenceID,
	}

	var resp payoutResponse
	path := fmt.Sprintf("/v2/programs/%s/payouts", g.programID)
	if err := g.do(ctx, http.MethodPost, path, payload, &resp); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{ExternalID: resp.PayoutID, Status: mapStatus(resp.Status), RawStatus: resp.Status}, nil
}

func (g *gateway) GetPayment(ctx context.Context, externalID string) (*port.GatewayPaymentResult, error) {
	var resp payoutResponse
	path := fmt.Sprintf("/v2/programs/%s/payouts/%s", g.programID, externalID)
	if err := g.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{ExternalID: resp.PayoutID, Status: mapStatus(resp.Status), RawStatus: resp.Status}, nil
}

func (g *gateway) RefundPayment(_ context.Context, _ port.RefundGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	return nil, port.ErrRefundNotSupported
}

func (g *gateway) ParseWebhook(_ context.Context, payload []byte, headers map[string][]string) (*port.WebhookEvent, error) {
	sig := http.Header(headers).Get("X-Payoneer-Signature")
	if sig == "" {
		return nil, fmt.Errorf("missing X-Payoneer-Signature header")
	}

	mac := hmac.New(sha256.New, []byte(g.webhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return nil, fmt.Errorf("signature mismatch")
	}

	var event struct {
		PayoutID string `json:"payout_id"`
		Status   string `json:"status"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode webhook payload: %w", err)
	}
	return &port.WebhookEvent{ExternalID: event.PayoutID, Status: mapStatus(event.Status), RawStatus: event.Status}, nil
}

func (g *gateway) do(ctx context.Context, method, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode payoneer request: %w", err)
		}
		body = strings.NewReader(string(raw))
	}

	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build payoneer request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("call payoneer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("payoneer returned status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode payoneer response: %w", err)
		}
	}
	return nil
}

func mapStatus(raw string) model.PaymentStatus {
	switch raw {
	case "COMPLETED", "PAID":
		return model.PaymentStatusCaptured
	case "PENDING", "PROCESSING":
		return model.PaymentStatusPending
	case "FAILED", "CANCELLED":
		return model.PaymentStatusFailed
	default:
		return model.PaymentStatusPending
	}
}
