package port

import (
	"context"

	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/google/uuid"
)

type CreateSubscriptionRequest struct {
	UserID          uuid.UUID
	PlanID          uuid.UUID
	Amount          float64
	Currency        string
	AutoRenewal     bool
	TrialDays       int
	PaymentMethodID uuid.UUID
}

type SubscriptionUsecase interface {
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*model.Subscription, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	CancelSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	ProcessDueRenewals(ctx context.Context) (int, error)
}

type AddPaymentMethodRequest struct {
	UserID         uuid.UUID
	Provider       string
	TokenID        string
	LastFourDigits string
	ExpiryDate     string
	IsDefault      bool
}

type PaymentMethodUsecase interface {
	AddPaymentMethod(ctx context.Context, req AddPaymentMethodRequest) (*model.PaymentMethod, error)
	ListPaymentMethods(ctx context.Context, userID uuid.UUID) ([]model.PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, id uuid.UUID) error
	SetDefaultPaymentMethod(ctx context.Context, userID, id uuid.UUID) error
}

type InvoiceUsecase interface {
	GetInvoice(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	ListInvoicesByUser(ctx context.Context, userID uuid.UUID) ([]model.Invoice, error)
	ListInvoicesBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Invoice, error)
}

type TransactionUsecase interface {
	ListTransactionsBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Transaction, error)
}
