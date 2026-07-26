package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	ListDueForRenewal(ctx context.Context, asOf time.Time) ([]model.Subscription, error)
}

type PaymentMethodRepository interface {
	Create(ctx context.Context, pm *model.PaymentMethod) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.PaymentMethod, error)
	GetDefaultByUser(ctx context.Context, userID uuid.UUID) (*model.PaymentMethod, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.PaymentMethod, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ClearDefaultByUser(ctx context.Context, userID uuid.UUID) error
	SetDefault(ctx context.Context, id uuid.UUID) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	ListBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Transaction, error)
}

type InvoiceRepository interface {
	Create(ctx context.Context, inv *model.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Invoice, error)
	ListBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Invoice, error)
}

type ChargeRequest struct {
	PaymentMethod *model.PaymentMethod
	Amount        float64
	Currency      string
	Description   string
}

type ChargeResult struct {
	Success              bool
	GatewayTransactionID string
	ErrorMessage         string
}

type PaymentGatewayPort interface {
	Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error)
}
