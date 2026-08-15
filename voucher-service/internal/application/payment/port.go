package payment

import (
	"context"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusRefunded  Status = "refunded"
)

type Payment struct {
	ID             string         `json:"id"`
	OrderID        shared.OrderID `json:"order_id"`
	Provider       string         `json:"provider"` // "vnpay" | "momo" | "wallet"
	Amount         shared.Money   `json:"amount"`
	Status         Status         `json:"status"`
	ProviderTxnRef string         `json:"provider_txn_ref,omitempty"`
}

type InitiatePaymentInput struct {
	OrderID        shared.OrderID
	Amount         shared.Money
	Provider       string
	ReturnURL      string
	IdempotencyKey string
}

type InitiatePaymentOutput struct {
	PaymentID   string `json:"payment_id"`
	RedirectURL string `json:"redirect_url"`
}

type WebhookInput struct {
	Provider  string
	RawBody   []byte
	Signature string
}

type PaymentService interface {
	InitiatePayment(ctx context.Context, in InitiatePaymentInput) (*InitiatePaymentOutput, error)
	HandleWebhook(ctx context.Context, in WebhookInput) error
	Refund(ctx context.Context, paymentID string, amount shared.Money) error
	GetPayment(ctx context.Context, id string) (*Payment, error)
	// ListSettledSince backs Reconciliation's PaymentRecordSource port via
	// an internalgateway adapter.
	ListSettledSince(ctx context.Context, since time.Time) ([]*Payment, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByProviderTxnRef(ctx context.Context, ref string) (*Payment, error)
	Save(ctx context.Context, p *Payment) error
	ListSettledSince(ctx context.Context, since time.Time) ([]*Payment, error)
}

type PaymentGateway interface {
	Provider() string
	InitPayment(ctx context.Context, refID string, amount shared.Money, returnURL string) (redirectURL string, err error)
	VerifyWebhookSignature(rawBody []byte, signature string) (providerTxnRef string, success bool, err error)
	Refund(ctx context.Context, providerTxnRef string, amount shared.Money) error
}

type PaymentGatewayRegistry interface {
	Resolve(provider string) (PaymentGateway, error)
}
