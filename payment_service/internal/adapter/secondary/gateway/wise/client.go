package wise

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/payment-service/config"
	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
)

type gateway struct {
	client           *http.Client
	baseURL          string
	apiToken         string
	profileID        string
	webhookPublicKey string
}

func NewGateway(cfg *config.Config) port.PaymentGateway {
	return &gateway{
		client:           &http.Client{Timeout: cfg.Wise.TimeoutDuration()},
		baseURL:          cfg.Wise.BaseURL,
		apiToken:         cfg.Wise.APIToken,
		profileID:        cfg.Wise.ProfileID,
		webhookPublicKey: cfg.Wise.WebhookPublicKey,
	}
}

func (g *gateway) Provider() model.Provider { return model.ProviderWise }

type quoteResponse struct {
	ID string `json:"id"`
}

type transferResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

func (g *gateway) CreatePayment(ctx context.Context, in port.CreateGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	recipientAccountID, ok := in.Metadata["recipient_account_id"]
	if !ok || recipientAccountID == "" {
		return nil, fmt.Errorf("wise: metadata[recipient_account_id] is required")
	}

	var quote quoteResponse
	quotePayload := map[string]any{
		"sourceCurrency": strings.ToUpper(in.Currency),
		"targetCurrency": strings.ToUpper(in.Currency),
		"sourceAmount":   float64(in.Amount) / 100,
	}
	quotePath := fmt.Sprintf("/v3/profiles/%s/quotes", g.profileID)
	if err := g.do(ctx, http.MethodPost, quotePath, quotePayload, &quote); err != nil {
		return nil, fmt.Errorf("create wise quote: %w", err)
	}

	var transfer transferResponse
	transferPayload := map[string]any{
		"targetAccount":         recipientAccountID,
		"quoteUuid":             quote.ID,
		"customerTransactionId": in.ReferenceID,
		"details": map[string]any{
			"reference": in.Description,
		},
	}
	if err := g.do(ctx, http.MethodPost, "/v1/transfers", transferPayload, &transfer); err != nil {
		return nil, fmt.Errorf("create wise transfer: %w", err)
	}

	externalID := fmt.Sprintf("%d", transfer.ID)
	return &port.GatewayPaymentResult{ExternalID: externalID, Status: mapStatus(transfer.Status), RawStatus: transfer.Status}, nil
}

func (g *gateway) GetPayment(ctx context.Context, externalID string) (*port.GatewayPaymentResult, error) {
	var resp transferResponse
	if err := g.do(ctx, http.MethodGet, "/v1/transfers/"+externalID, nil, &resp); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{ExternalID: externalID, Status: mapStatus(resp.Status), RawStatus: resp.Status}, nil
}

func (g *gateway) RefundPayment(ctx context.Context, in port.RefundGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	if in.Amount != in.OutstandingBalance {
		return nil, port.ErrRefundNotSupported
	}
	if err := g.do(ctx, http.MethodPut, "/v1/transfers/"+in.ExternalID+"/cancel", nil, nil); err != nil {
		return nil, err
	}
	return &port.GatewayPaymentResult{ExternalID: in.ExternalID, Status: model.PaymentStatusRefunded}, nil
}

func (g *gateway) ParseWebhook(_ context.Context, payload []byte, headers map[string][]string) (*port.WebhookEvent, error) {
	sigB64 := http.Header(headers).Get("X-Signature-SHA256")
	if sigB64 == "" {
		return nil, fmt.Errorf("missing X-Signature-SHA256 header")
	}
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	pub, err := parseRSAPublicKey(g.webhookPublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse configured public key: %w", err)
	}
	hashed := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig); err != nil {
		return nil, fmt.Errorf("signature mismatch: %w", err)
	}

	var event struct {
		Data struct {
			Resource struct {
				ID int64 `json:"id"`
			} `json:"resource"`
			CurrentState string `json:"current_state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode webhook payload: %w", err)
	}

	externalID := fmt.Sprintf("%d", event.Data.Resource.ID)
	return &port.WebhookEvent{ExternalID: externalID, Status: mapStatus(event.Data.CurrentState), RawStatus: event.Data.CurrentState}, nil
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("configured key is not an RSA public key")
	}
	return rsaKey, nil
}

func (g *gateway) do(ctx context.Context, method, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode wise request: %w", err)
		}
		body = strings.NewReader(string(raw))
	}

	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build wise request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("call wise: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("wise returned status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode wise response: %w", err)
		}
	}
	return nil
}

func mapStatus(raw string) model.PaymentStatus {
	switch raw {
	case "outgoing_payment_sent", "funds_converted", "completed":
		return model.PaymentStatusCaptured
	case "processing", "incoming_payment_waiting":
		return model.PaymentStatusPending
	case "cancelled", "funds_refunded", "bounced_back":
		return model.PaymentStatusFailed
	default:
		return model.PaymentStatusPending
	}
}
