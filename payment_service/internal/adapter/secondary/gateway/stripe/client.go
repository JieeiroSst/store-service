package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/payment-service/config"
	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
)

type gateway struct {
	client        *http.Client
	baseURL       string
	secretKey     string
	webhookSecret string
}

func NewGateway(cfg *config.Config) port.PaymentGateway {
	return &gateway{
		client:        &http.Client{Timeout: cfg.Stripe.TimeoutDuration()},
		baseURL:       cfg.Stripe.BaseURL,
		secretKey:     cfg.Stripe.SecretKey,
		webhookSecret: cfg.Stripe.WebhookSecret,
	}
}

func (g *gateway) Provider() model.Provider { return model.ProviderStripe }

type paymentIntentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (g *gateway) CreatePayment(ctx context.Context, in port.CreateGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	form := url.Values{}
	form.Set("amount", strconv.FormatInt(in.Amount, 10))
	form.Set("currency", strings.ToLower(in.Currency))
	form.Add("payment_method_types[]", "card")
	if in.Description != "" {
		form.Set("description", in.Description)
	}
	if in.PayerEmail != "" {
		form.Set("receipt_email", in.PayerEmail)
	}
	if in.ReferenceID != "" {
		form.Set("metadata[reference_id]", in.ReferenceID)
	}

	var resp paymentIntentResponse
	if err := g.do(ctx, http.MethodPost, "/v1/payment_intents", form, &resp); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{
		ExternalID: resp.ID,
		Status:     mapStatus(resp.Status),
		RawStatus:  resp.Status,
	}, nil
}

func (g *gateway) GetPayment(ctx context.Context, externalID string) (*port.GatewayPaymentResult, error) {
	var resp paymentIntentResponse
	if err := g.do(ctx, http.MethodGet, "/v1/payment_intents/"+externalID, nil, &resp); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{ExternalID: resp.ID, Status: mapStatus(resp.Status), RawStatus: resp.Status}, nil
}

func (g *gateway) RefundPayment(ctx context.Context, in port.RefundGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	form := url.Values{}
	form.Set("payment_intent", in.ExternalID)
	form.Set("amount", strconv.FormatInt(in.Amount, 10))

	var resp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := g.do(ctx, http.MethodPost, "/v1/refunds", form, &resp); err != nil {
		return nil, err
	}
	status := model.PaymentStatusRefunded
	if in.Amount < in.OutstandingBalance {
		status = model.PaymentStatusPartiallyRefunded
	}
	return &port.GatewayPaymentResult{
		ExternalID: in.ExternalID,
		Status:     status,
		RawStatus:  resp.Status,
	}, nil
}

// ParseWebhook verifies the Stripe-Signature header
// (https://stripe.com/docs/webhooks/signatures: "t=<timestamp>,v1=<hex hmac>")
// by recomputing the HMAC-SHA256 of "<timestamp>.<payload>" with webhookSecret,
// then parses the event for the payment intent id and status.
func (g *gateway) ParseWebhook(_ context.Context, payload []byte, headers map[string][]string) (*port.WebhookEvent, error) {
	sigHeader := http.Header(headers).Get("Stripe-Signature")
	if sigHeader == "" {
		return nil, fmt.Errorf("missing Stripe-Signature header")
	}

	var timestamp, signature string
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signature = kv[1]
		}
	}
	if timestamp == "" || signature == "" {
		return nil, fmt.Errorf("malformed Stripe-Signature header")
	}

	mac := hmac.New(sha256.New, []byte(g.webhookSecret))
	mac.Write([]byte(timestamp + "." + string(payload)))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return nil, fmt.Errorf("signature mismatch")
	}

	var event struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID            string `json:"id"`
				Status        string `json:"status"`
				PaymentIntent string `json:"payment_intent"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode webhook payload: %w", err)
	}

	externalID := event.Data.Object.ID
	status := mapStatus(event.Data.Object.Status)
	if event.Type == "charge.refunded" {
		externalID = event.Data.Object.PaymentIntent
		status = model.PaymentStatusRefunded
	}
	return &port.WebhookEvent{ExternalID: externalID, Status: status, RawStatus: event.Data.Object.Status}, nil
}

func (g *gateway) do(ctx context.Context, method, path string, form url.Values, out any) error {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build stripe request: %w", err)
	}
	req.SetBasicAuth(g.secretKey, "")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("call stripe: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("stripe returned status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode stripe response: %w", err)
		}
	}
	return nil
}

func mapStatus(raw string) model.PaymentStatus {
	switch raw {
	case "succeeded":
		return model.PaymentStatusCaptured
	case "requires_capture":
		return model.PaymentStatusAuthorized
	case "canceled":
		return model.PaymentStatusFailed
	default:
		return model.PaymentStatusPending
	}
}
