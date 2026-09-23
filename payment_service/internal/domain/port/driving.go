package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
)

type CreatePaymentInput struct {
	Provider       model.Provider
	Amount         int64
	Currency       string
	PayerEmail     string
	Description    string
	Metadata       map[string]string
	IdempotencyKey string
}

type ListPaymentsInput struct {
	Provider model.Provider
	Status   model.PaymentStatus
	From     time.Time
	To       time.Time
	Limit    int
	Offset   int
}

type PaymentList struct {
	Items []model.Payment
	Total int64
}

type PaymentUsecase interface {
	CreatePayment(ctx context.Context, in CreatePaymentInput) (*model.Payment, error)
	GetPayment(ctx context.Context, id int64) (*model.Payment, error)
	ListPayments(ctx context.Context, in ListPaymentsInput) (*PaymentList, error)
	RefundPayment(ctx context.Context, id int64, amount int64) (*model.Payment, error)
	ListTransactions(ctx context.Context, paymentID int64) ([]model.Transaction, error)
	HandleWebhook(ctx context.Context, provider model.Provider, payload []byte, headers map[string][]string) error
}
