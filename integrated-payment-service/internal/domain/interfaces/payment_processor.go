package interfaces

import (
	"context"

	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/entities"
)

type PaymentProcessor interface {
	CreatePayment(ctx context.Context, payment *entities.Payment) (*PaymentResponse, error)
	ProcessPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
	RefundPayment(ctx context.Context, paymentID string, amount float64) (*PaymentResponse, error)
	GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentResponse, error)
	ValidateWebhook(ctx context.Context, payload []byte, signature string) error
}

type PaymentResponse struct {
	ID            string                 `json:"id"`
	Status        entities.PaymentStatus `json:"status"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	PaymentURL    string                 `json:"payment_url,omitempty"`
	TransactionID string                 `json:"transaction_id,omitempty"`
	ProcessorData map[string]interface{} `json:"processor_data,omitempty"`
	Message       string                 `json:"message,omitempty"`
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *entities.Payment) error
	GetByID(ctx context.Context, id string) (*entities.Payment, error)
	Update(ctx context.Context, payment *entities.Payment) error
	GetByUserID(ctx context.Context, userID string) ([]*entities.Payment, error)
}
