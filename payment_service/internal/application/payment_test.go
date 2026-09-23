package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
)

func TestCreatePayment_Success(t *testing.T) {
	repo := newFakeRepository()
	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderStripe: &fakeGateway{
			provider: model.ProviderStripe,
			result:   &port.GatewayPaymentResult{ExternalID: "pi_123", Status: model.PaymentStatusCaptured},
		},
	}}
	svc := NewPaymentService(repo, newFakeTransactionRepository(), resolver)

	payment, err := svc.CreatePayment(context.Background(), port.CreatePaymentInput{
		Provider: model.ProviderStripe,
		Amount:   1000,
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != model.PaymentStatusCaptured {
		t.Fatalf("expected status captured, got %s", payment.Status)
	}
	if payment.ExternalID != "pi_123" {
		t.Fatalf("expected external id pi_123, got %s", payment.ExternalID)
	}
}

func TestCreatePayment_UnsupportedProvider(t *testing.T) {
	svc := NewPaymentService(newFakeRepository(), newFakeTransactionRepository(), &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{}})

	_, err := svc.CreatePayment(context.Background(), port.CreatePaymentInput{
		Provider: model.Provider("unknown"),
		Amount:   1000,
		Currency: "USD",
	})
	if !errors.Is(err, port.ErrInvalidProvider) {
		t.Fatalf("expected ErrInvalidProvider, got %v", err)
	}
}

func TestCreatePayment_GatewayFailureMarksPaymentFailed(t *testing.T) {
	repo := newFakeRepository()
	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderPayPal: &fakeGateway{provider: model.ProviderPayPal, err: errors.New("boom")},
	}}
	svc := NewPaymentService(repo, newFakeTransactionRepository(), resolver)

	payment, err := svc.CreatePayment(context.Background(), port.CreatePaymentInput{
		Provider: model.ProviderPayPal,
		Amount:   500,
		Currency: "USD",
	})
	if !errors.Is(err, port.ErrGatewayRequestFailed) {
		t.Fatalf("expected ErrGatewayRequestFailed, got %v", err)
	}
	if payment == nil || payment.Status != model.PaymentStatusFailed {
		t.Fatalf("expected persisted payment with status failed, got %+v", payment)
	}

	stored, err := repo.GetByID(context.Background(), payment.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching stored payment: %v", err)
	}
	if stored.Status != model.PaymentStatusFailed {
		t.Fatalf("expected stored status failed, got %s", stored.Status)
	}
}

func TestRefundPayment_RejectsNonCaptured(t *testing.T) {
	repo := newFakeRepository()
	created, err := repo.Create(context.Background(), &model.Payment{
		Provider: model.ProviderStripe,
		Amount:   1000,
		Currency: "USD",
		Status:   model.PaymentStatusPending,
	})
	if err != nil {
		t.Fatalf("unexpected error seeding payment: %v", err)
	}

	svc := NewPaymentService(repo, newFakeTransactionRepository(), &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{}})

	_, err = svc.RefundPayment(context.Background(), created.ID, 0)
	if !errors.Is(err, port.ErrPaymentNotRefundable) {
		t.Fatalf("expected ErrPaymentNotRefundable, got %v", err)
	}
}

func TestRefundPayment_Success(t *testing.T) {
	repo := newFakeRepository()
	created, err := repo.Create(context.Background(), &model.Payment{
		Provider:   model.ProviderStripe,
		Amount:     1000,
		Currency:   "USD",
		Status:     model.PaymentStatusCaptured,
		ExternalID: "pi_123",
	})
	if err != nil {
		t.Fatalf("unexpected error seeding payment: %v", err)
	}

	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderStripe: &fakeGateway{
			provider: model.ProviderStripe,
			result:   &port.GatewayPaymentResult{ExternalID: "pi_123", Status: model.PaymentStatusRefunded},
		},
	}}
	svc := NewPaymentService(repo, newFakeTransactionRepository(), resolver)

	payment, err := svc.RefundPayment(context.Background(), created.ID, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != model.PaymentStatusRefunded {
		t.Fatalf("expected status refunded, got %s", payment.Status)
	}
}

func TestListTransactions_RecordsLifecycle(t *testing.T) {
	repo := newFakeRepository()
	transactions := newFakeTransactionRepository()
	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderStripe: &fakeGateway{
			provider: model.ProviderStripe,
			result:   &port.GatewayPaymentResult{ExternalID: "pi_123", Status: model.PaymentStatusCaptured},
		},
	}}
	svc := NewPaymentService(repo, transactions, resolver)

	created, err := svc.CreatePayment(context.Background(), port.CreatePaymentInput{
		Provider: model.ProviderStripe,
		Amount:   1000,
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error creating payment: %v", err)
	}

	resolver.gateways[model.ProviderStripe] = &fakeGateway{
		provider: model.ProviderStripe,
		result:   &port.GatewayPaymentResult{ExternalID: "pi_123", Status: model.PaymentStatusRefunded},
	}
	if _, err := svc.RefundPayment(context.Background(), created.ID, 0); err != nil {
		t.Fatalf("unexpected error refunding payment: %v", err)
	}

	log, err := svc.ListTransactions(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error listing transactions: %v", err)
	}
	if len(log) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(log))
	}
	if log[0].Type != model.TransactionTypeCreate || log[0].Status != model.PaymentStatusCaptured {
		t.Fatalf("expected first transaction to be a captured create, got %+v", log[0])
	}
	if log[1].Type != model.TransactionTypeRefund || log[1].Status != model.PaymentStatusRefunded {
		t.Fatalf("expected second transaction to be a refunded refund, got %+v", log[1])
	}
}

func TestListTransactions_UnknownPayment(t *testing.T) {
	svc := NewPaymentService(newFakeRepository(), newFakeTransactionRepository(), &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{}})

	_, err := svc.ListTransactions(context.Background(), 999)
	if !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreatePayment_IdempotentReplayDoesNotCallGatewayAgain(t *testing.T) {
	repo := newFakeRepository()
	calls := 0
	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderStripe: &countingGateway{
			fakeGateway: fakeGateway{
				provider: model.ProviderStripe,
				result:   &port.GatewayPaymentResult{ExternalID: "pi_123", Status: model.PaymentStatusCaptured},
			},
			calls: &calls,
		},
	}}
	svc := NewPaymentService(repo, newFakeTransactionRepository(), resolver)

	in := port.CreatePaymentInput{
		Provider:       model.ProviderStripe,
		Amount:         1000,
		Currency:       "USD",
		IdempotencyKey: "key-1",
	}

	first, err := svc.CreatePayment(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	second, err := svc.CreatePayment(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected replay to return the same payment, got %d and %d", first.ID, second.ID)
	}
	if calls != 1 {
		t.Fatalf("expected the gateway to be called once, got %d", calls)
	}
}

func TestRefundPayment_PartialThenFull(t *testing.T) {
	repo := newFakeRepository()
	created, err := repo.Create(context.Background(), &model.Payment{
		Provider:   model.ProviderStripe,
		Amount:     1000,
		Currency:   "USD",
		Status:     model.PaymentStatusCaptured,
		ExternalID: "pi_123",
	})
	if err != nil {
		t.Fatalf("unexpected error seeding payment: %v", err)
	}

	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderStripe: &fakeGateway{
			provider: model.ProviderStripe,
			result:   &port.GatewayPaymentResult{ExternalID: "pi_123", Status: model.PaymentStatusPartiallyRefunded},
		},
	}}
	svc := NewPaymentService(repo, newFakeTransactionRepository(), resolver)

	payment, err := svc.RefundPayment(context.Background(), created.ID, 400)
	if err != nil {
		t.Fatalf("unexpected error on first refund: %v", err)
	}
	if payment.Status != model.PaymentStatusPartiallyRefunded {
		t.Fatalf("expected partially_refunded after first refund, got %s", payment.Status)
	}

	payment, err = svc.RefundPayment(context.Background(), created.ID, 0)
	if err != nil {
		t.Fatalf("unexpected error on second refund: %v", err)
	}
	if payment.Status != model.PaymentStatusRefunded {
		t.Fatalf("expected refunded after clearing the balance, got %s", payment.Status)
	}
	if payment.RefundedAmount != 1000 {
		t.Fatalf("expected refunded_amount 1000, got %d", payment.RefundedAmount)
	}
}

func TestRefundPayment_RejectsOverAmount(t *testing.T) {
	repo := newFakeRepository()
	created, err := repo.Create(context.Background(), &model.Payment{
		Provider:   model.ProviderStripe,
		Amount:     1000,
		Currency:   "USD",
		Status:     model.PaymentStatusCaptured,
		ExternalID: "pi_123",
	})
	if err != nil {
		t.Fatalf("unexpected error seeding payment: %v", err)
	}

	svc := NewPaymentService(repo, newFakeTransactionRepository(), &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{}})

	_, err = svc.RefundPayment(context.Background(), created.ID, 1500)
	if !errors.Is(err, port.ErrInvalidRefundAmount) {
		t.Fatalf("expected ErrInvalidRefundAmount, got %v", err)
	}
}

func TestListPayments_FiltersByProvider(t *testing.T) {
	repo := newFakeRepository()
	if _, err := repo.Create(context.Background(), &model.Payment{Provider: model.ProviderStripe, Amount: 100, Currency: "USD"}); err != nil {
		t.Fatalf("unexpected error seeding stripe payment: %v", err)
	}
	if _, err := repo.Create(context.Background(), &model.Payment{Provider: model.ProviderPayPal, Amount: 200, Currency: "USD"}); err != nil {
		t.Fatalf("unexpected error seeding paypal payment: %v", err)
	}

	svc := NewPaymentService(repo, newFakeTransactionRepository(), &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{}})

	result, err := svc.ListPayments(context.Background(), port.ListPaymentsInput{Provider: model.ProviderStripe})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("expected exactly 1 stripe payment, got total=%d items=%d", result.Total, len(result.Items))
	}
	if result.Items[0].Provider != model.ProviderStripe {
		t.Fatalf("expected stripe payment, got %s", result.Items[0].Provider)
	}
}

func TestHandleWebhook_UpdatesPaymentAndLogsTransaction(t *testing.T) {
	repo := newFakeRepository()
	transactions := newFakeTransactionRepository()
	created, err := repo.Create(context.Background(), &model.Payment{
		Provider:   model.ProviderStripe,
		Amount:     1000,
		Currency:   "USD",
		Status:     model.PaymentStatusPending,
		ExternalID: "pi_123",
	})
	if err != nil {
		t.Fatalf("unexpected error seeding payment: %v", err)
	}

	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderStripe: &fakeGateway{
			provider:     model.ProviderStripe,
			webhookEvent: &port.WebhookEvent{ExternalID: "pi_123", Status: model.PaymentStatusCaptured, RawStatus: "succeeded"},
		},
	}}
	svc := NewPaymentService(repo, transactions, resolver)

	if err := svc.HandleWebhook(context.Background(), model.ProviderStripe, []byte(`{}`), map[string][]string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching payment: %v", err)
	}
	if updated.Status != model.PaymentStatusCaptured {
		t.Fatalf("expected status captured, got %s", updated.Status)
	}

	log, err := svc.ListTransactions(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error listing transactions: %v", err)
	}
	if len(log) != 1 || log[0].Type != model.TransactionTypeWebhook {
		t.Fatalf("expected a single webhook transaction, got %+v", log)
	}
}

func TestHandleWebhook_RejectsInvalidSignature(t *testing.T) {
	resolver := &fakeResolver{gateways: map[model.Provider]port.PaymentGateway{
		model.ProviderStripe: &fakeGateway{provider: model.ProviderStripe, webhookErr: errors.New("bad signature")},
	}}
	svc := NewPaymentService(newFakeRepository(), newFakeTransactionRepository(), resolver)

	err := svc.HandleWebhook(context.Background(), model.ProviderStripe, []byte(`{}`), map[string][]string{})
	if !errors.Is(err, port.ErrInvalidWebhook) {
		t.Fatalf("expected ErrInvalidWebhook, got %v", err)
	}
}

// countingGateway wraps fakeGateway to count CreatePayment calls, used to
// prove an idempotent replay doesn't hit the gateway a second time.
type countingGateway struct {
	fakeGateway
	calls *int
}

func (g *countingGateway) CreatePayment(ctx context.Context, in port.CreateGatewayPaymentInput) (*port.GatewayPaymentResult, error) {
	*g.calls++
	return g.fakeGateway.CreatePayment(ctx, in)
}
