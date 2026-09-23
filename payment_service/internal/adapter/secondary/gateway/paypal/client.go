package paypal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/payment-service/config"
	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
)

type gateway struct {
	client       *http.Client
	baseURL      string
	clientID     string
	clientSecret string
	webhookID    string

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func NewGateway(cfg *config.Config) port.PaymentGateway {
	return &gateway{
		client:       &http.Client{Timeout: cfg.PayPal.TimeoutDuration()},
		baseURL:      cfg.PayPal.BaseURL,
		clientID:     cfg.PayPal.ClientID,
		clientSecret: cfg.PayPal.ClientSecret,
		webhookID:    cfg.PayPal.WebhookID,
	}
}

func (g *gateway) Provider() model.Provider { return model.ProviderPayPal }

type orderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (g *gateway) CreatePayment(ctx context.Context, in port.CreateGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	token, err := g.token(ctx)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{
			{
				"reference_id": in.ReferenceID,
				"description":  in.Description,
				"amount": map[string]any{
					"currency_code": strings.ToUpper(in.Currency),
					"value":         minorUnitsToDecimalString(in.Amount),
				},
			},
		},
	}

	var resp orderResponse
	if err := g.do(ctx, http.MethodPost, "/v2/checkout/orders", token, payload, &resp); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{ExternalID: resp.ID, Status: mapStatus(resp.Status), RawStatus: resp.Status}, nil
}

func (g *gateway) GetPayment(ctx context.Context, externalID string) (*port.GatewayPaymentResult, error) {
	token, err := g.token(ctx)
	if err != nil {
		return nil, err
	}

	var resp orderResponse
	if err := g.do(ctx, http.MethodGet, "/v2/checkout/orders/"+externalID, token, nil, &resp); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{ExternalID: resp.ID, Status: mapStatus(resp.Status), RawStatus: resp.Status}, nil
}

func (g *gateway) RefundPayment(ctx context.Context, in port.RefundGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	token, err := g.token(ctx)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"amount": map[string]any{
			"currency_code": strings.ToUpper(in.Currency),
			"value":         minorUnitsToDecimalString(in.Amount),
		},
	}
	var resp orderResponse
	path := fmt.Sprintf("/v2/payments/captures/%s/refund", in.ExternalID)
	if err := g.do(ctx, http.MethodPost, path, token, payload, &resp); err != nil {
		return nil, err
	}
	status := model.PaymentStatusRefunded
	if in.Amount < in.OutstandingBalance {
		status = model.PaymentStatusPartiallyRefunded
	}
	return &port.GatewayPaymentResult{ExternalID: in.ExternalID, Status: status, RawStatus: resp.Status}, nil
}

func (g *gateway) ParseWebhook(ctx context.Context, payload []byte, headers map[string][]string) (*port.WebhookEvent, error) {
	token, err := g.token(ctx)
	if err != nil {
		return nil, err
	}

	h := http.Header(headers)
	var rawEvent map[string]any
	if err := json.Unmarshal(payload, &rawEvent); err != nil {
		return nil, fmt.Errorf("decode webhook payload: %w", err)
	}

	verifyPayload := map[string]any{
		"auth_algo":         h.Get("Paypal-Auth-Algo"),
		"cert_url":          h.Get("Paypal-Cert-Url"),
		"transmission_id":   h.Get("Paypal-Transmission-Id"),
		"transmission_sig":  h.Get("Paypal-Transmission-Sig"),
		"transmission_time": h.Get("Paypal-Transmission-Time"),
		"webhook_id":        g.webhookID,
		"webhook_event":     rawEvent,
	}
	var verifyResp struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := g.do(ctx, http.MethodPost, "/v1/notifications/verify-webhook-signature", token, verifyPayload, &verifyResp); err != nil {
		return nil, fmt.Errorf("call verify-webhook-signature: %w", err)
	}
	if verifyResp.VerificationStatus != "SUCCESS" {
		return nil, fmt.Errorf("webhook signature verification status %q", verifyResp.VerificationStatus)
	}

	var event struct {
		EventType string `json:"event_type"`
		Resource  struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode webhook event: %w", err)
	}

	return &port.WebhookEvent{ExternalID: event.Resource.ID, Status: mapStatus(event.Resource.Status), RawStatus: event.Resource.Status}, nil
}

func (g *gateway) token(ctx context.Context) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.accessToken != "" && time.Now().Before(g.expiresAt.Add(-time.Minute)) {
		return g.accessToken, nil
	}

	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/v1/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build paypal token request: %w", err)
	}
	auth := base64.StdEncoding.EncodeToString([]byte(g.clientID + ":" + g.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call paypal oauth: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("paypal oauth returned status %d", resp.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode paypal oauth response: %w", err)
	}

	g.accessToken = body.AccessToken
	g.expiresAt = time.Now().Add(time.Duration(body.ExpiresIn) * time.Second)
	return g.accessToken, nil
}

func (g *gateway) do(ctx context.Context, method, path, token string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode paypal request: %w", err)
		}
		body = strings.NewReader(string(raw))
	}

	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build paypal request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("call paypal: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("paypal returned status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode paypal response: %w", err)
		}
	}
	return nil
}

func mapStatus(raw string) model.PaymentStatus {
	switch raw {
	case "COMPLETED":
		return model.PaymentStatusCaptured
	case "APPROVED":
		return model.PaymentStatusAuthorized
	case "VOIDED":
		return model.PaymentStatusFailed
	default:
		return model.PaymentStatusPending
	}
}

func minorUnitsToDecimalString(amount int64) string {
	return strconv.FormatFloat(float64(amount)/100, 'f', 2, 64)
}
