package payments

import (
	"context"

	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/entities"
	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/interfaces"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	ReturnURL     string
}

type StripeProcessor struct {
	config StripeConfig
}

func NewStripeProcessor(config StripeConfig) *StripeProcessor {
	stripe.Key = config.SecretKey
	return &StripeProcessor{config: config}
}

func (s *StripeProcessor) CreatePayment(ctx context.Context, payment *entities.Payment) (*interfaces.PaymentResponse, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(payment.Amount * 100)), // Convert to cents
		Currency: stripe.String(payment.Currency),
		Metadata: map[string]string{
			"payment_id": payment.ID,
			"user_id":    payment.UserID,
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	status := entities.PaymentStatusPending
	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded:
		status = entities.PaymentStatusSuccess
	case stripe.PaymentIntentStatusProcessing:
		status = entities.PaymentStatusProcessing
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		status = entities.PaymentStatusPending
	}

	return &interfaces.PaymentResponse{
		ID:            payment.ID,
		Status:        status,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		TransactionID: pi.ID,
		ProcessorData: map[string]interface{}{
			"client_secret":     pi.ClientSecret,
			"payment_intent_id": pi.ID,
		},
	}, nil
}

func (s *StripeProcessor) ProcessPayment(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (s *StripeProcessor) RefundPayment(ctx context.Context, paymentID string, amount float64) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (s *StripeProcessor) GetPaymentStatus(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (s *StripeProcessor) ValidateWebhook(ctx context.Context, payload []byte, signature string) error {
	return nil
}
