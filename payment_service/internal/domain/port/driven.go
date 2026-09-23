package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
)

type ListPaymentsFilter struct {
	Provider model.Provider
	Status   model.PaymentStatus
	From     time.Time
	To       time.Time
	Limit    int
	Offset   int
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) (*model.Payment, error)
	GetByID(ctx context.Context, id int64) (*model.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*model.Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Payment, error)
	List(ctx context.Context, filter ListPaymentsFilter) ([]model.Payment, int64, error)
	UpdateStatus(ctx context.Context, id int64, status model.PaymentStatus, externalID, failureReason string) error
	ApplyRefund(ctx context.Context, id int64, refundedDelta int64, status model.PaymentStatus) error
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error)
	ListByPaymentID(ctx context.Context, paymentID int64) ([]model.Transaction, error)
}

type CreateGatewayPaymentInput struct {
	Amount      int64
	Currency    string
	PayerEmail  string
	Description string
	ReferenceID string
	Metadata    map[string]string
}

type GatewayPaymentResult struct {
	ExternalID string
	Status     model.PaymentStatus
	RawStatus  string
}

type RefundGatewayPaymentInput struct {
	ExternalID         string
	Amount             int64
	Currency           string
	OutstandingBalance int64
}

type WebhookEvent struct {
	ExternalID string
	Status     model.PaymentStatus
	RawStatus  string
}

type PaymentGateway interface {
	Provider() model.Provider
	CreatePayment(ctx context.Context, in CreateGatewayPaymentInput) (*GatewayPaymentResult, error)
	GetPayment(ctx context.Context, externalID string) (*GatewayPaymentResult, error)
	RefundPayment(ctx context.Context, in RefundGatewayPaymentInput) (*GatewayPaymentResult, error)
	ParseWebhook(ctx context.Context, payload []byte, headers map[string][]string) (*WebhookEvent, error)
}

type PaymentGatewayResolver interface {
	Resolve(provider model.Provider) (PaymentGateway, error)
}
