package application

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
)

type fakeRepository struct {
	payments map[int64]*model.Payment
	nextID   int64
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{payments: make(map[int64]*model.Payment)}
}

func (r *fakeRepository) Create(_ context.Context, payment *model.Payment) (*model.Payment, error) {
	r.nextID++
	payment.ID = r.nextID
	clone := *payment
	r.payments[payment.ID] = &clone
	return &clone, nil
}

func (r *fakeRepository) GetByID(_ context.Context, id int64) (*model.Payment, error) {
	payment, ok := r.payments[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	clone := *payment
	return &clone, nil
}

func (r *fakeRepository) GetByExternalID(_ context.Context, externalID string) (*model.Payment, error) {
	for _, payment := range r.payments {
		if payment.ExternalID == externalID {
			clone := *payment
			return &clone, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r *fakeRepository) GetByIdempotencyKey(_ context.Context, key string) (*model.Payment, error) {
	for _, payment := range r.payments {
		if payment.IdempotencyKey != "" && payment.IdempotencyKey == key {
			clone := *payment
			return &clone, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r *fakeRepository) List(_ context.Context, filter port.ListPaymentsFilter) ([]model.Payment, int64, error) {
	var matched []model.Payment
	for _, payment := range r.payments {
		if filter.Provider != "" && payment.Provider != filter.Provider {
			continue
		}
		if filter.Status != "" && payment.Status != filter.Status {
			continue
		}
		matched = append(matched, *payment)
	}
	total := int64(len(matched))

	limit := filter.Limit
	if limit <= 0 || filter.Offset+limit > len(matched) {
		limit = len(matched) - filter.Offset
	}
	if filter.Offset >= len(matched) {
		return []model.Payment{}, total, nil
	}
	return matched[filter.Offset : filter.Offset+limit], total, nil
}

func (r *fakeRepository) ApplyRefund(_ context.Context, id int64, refundedDelta int64, status model.PaymentStatus) error {
	payment, ok := r.payments[id]
	if !ok {
		return port.ErrNotFound
	}
	payment.RefundedAmount += refundedDelta
	payment.Status = status
	return nil
}

func (r *fakeRepository) UpdateStatus(_ context.Context, id int64, status model.PaymentStatus, externalID, failureReason string) error {
	payment, ok := r.payments[id]
	if !ok {
		return port.ErrNotFound
	}
	payment.Status = status
	if externalID != "" {
		payment.ExternalID = externalID
	}
	payment.FailureReason = failureReason
	return nil
}

type fakeTransactionRepository struct {
	transactions []model.Transaction
	nextID       int64
}

func newFakeTransactionRepository() *fakeTransactionRepository {
	return &fakeTransactionRepository{}
}

func (r *fakeTransactionRepository) Create(_ context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	r.nextID++
	transaction.ID = r.nextID
	r.transactions = append(r.transactions, *transaction)
	return transaction, nil
}

func (r *fakeTransactionRepository) ListByPaymentID(_ context.Context, paymentID int64) ([]model.Transaction, error) {
	var result []model.Transaction
	for _, tx := range r.transactions {
		if tx.PaymentID == paymentID {
			result = append(result, tx)
		}
	}
	return result, nil
}

type fakeGateway struct {
	provider     model.Provider
	result       *port.GatewayPaymentResult
	err          error
	webhookEvent *port.WebhookEvent
	webhookErr   error
}

func (g *fakeGateway) Provider() model.Provider { return g.provider }

func (g *fakeGateway) CreatePayment(_ context.Context, _ port.CreateGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.result, nil
}

func (g *fakeGateway) GetPayment(_ context.Context, _ string) (*port.GatewayPaymentResult, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.result, nil
}

func (g *fakeGateway) RefundPayment(_ context.Context, _ port.RefundGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.result, nil
}

func (g *fakeGateway) ParseWebhook(_ context.Context, _ []byte, _ map[string][]string) (*port.WebhookEvent, error) {
	if g.webhookErr != nil {
		return nil, g.webhookErr
	}
	return g.webhookEvent, nil
}

type fakeResolver struct {
	gateways map[model.Provider]port.PaymentGateway
}

func (r *fakeResolver) Resolve(provider model.Provider) (port.PaymentGateway, error) {
	gateway, ok := r.gateways[provider]
	if !ok {
		return nil, errors.New("no gateway registered for provider")
	}
	return gateway, nil
}
