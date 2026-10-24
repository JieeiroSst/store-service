package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"go.uber.org/zap"
)

type memRepo struct {
	PaymentRepository
	byID map[string]*Payment
}

func (r *memRepo) FindByID(_ context.Context, id string) (*Payment, error) {
	p, ok := r.byID[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *memRepo) Save(_ context.Context, p *Payment) error {
	cp := *p
	r.byID[p.ID] = &cp
	return nil
}

type stubGateway struct {
	PaymentGateway
	name string
	res  WebhookResult
	err  error
}

func (g stubGateway) Provider() string { return g.name }

func (g stubGateway) VerifyWebhookSignature([]byte, string) (WebhookResult, error) {
	return g.res, g.err
}

type stubRegistry struct{ g PaymentGateway }

func (r stubRegistry) Resolve(string) (PaymentGateway, error) { return r.g, nil }

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type countingOutbox struct{ n int }

func (o *countingOutbox) Enqueue(context.Context, outbox.Event) error { o.n++; return nil }

func newSvc(g stubGateway, p Payment) (*Service, *memRepo, *countingOutbox) {
	repo := &memRepo{byID: map[string]*Payment{p.ID: &p}}
	ob := &countingOutbox{}
	return NewService(repo, stubRegistry{g}, inlineTx{}, ob, zap.NewNop()).(*Service), repo, ob
}

func pending() Payment {
	return Payment{ID: "pay-1", Provider: "momo", Amount: shared.Money{Amount: 50000, Currency: "VND"}, Status: StatusPending}
}

func TestWebhookRejectsBadSignature(t *testing.T) {
	svc, repo, ob := newSvc(stubGateway{name: "momo", err: errors.New("signature mismatch")}, pending())
	if err := svc.HandleWebhook(context.Background(), WebhookInput{Provider: "momo"}); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if repo.byID["pay-1"].Status != StatusPending || ob.n != 0 {
		t.Fatal("payment changed by a forged webhook")
	}
}

func TestWebhookRejectsAmountMismatch(t *testing.T) {
	svc, repo, ob := newSvc(stubGateway{name: "momo", res: WebhookResult{PaymentID: "pay-1", Success: true, Amount: 1}}, pending())
	if err := svc.HandleWebhook(context.Background(), WebhookInput{Provider: "momo"}); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if repo.byID["pay-1"].Status != StatusPending || ob.n != 0 {
		t.Fatal("underpaid payment was settled")
	}
}

func TestWebhookRejectsProviderMismatch(t *testing.T) {
	svc, _, ob := newSvc(stubGateway{name: "vnpay", res: WebhookResult{PaymentID: "pay-1", Success: true, Amount: 50000}}, pending())
	if err := svc.HandleWebhook(context.Background(), WebhookInput{Provider: "vnpay"}); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if ob.n != 0 {
		t.Fatal("settled through the wrong provider")
	}
}

func TestWebhookSettlesOnceAndIgnoresLaterFailure(t *testing.T) {
	ok := stubGateway{name: "momo", res: WebhookResult{PaymentID: "pay-1", Success: true, Amount: 50000}}
	svc, repo, ob := newSvc(ok, pending())
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if err := svc.HandleWebhook(ctx, WebhookInput{Provider: "momo"}); err != nil {
			t.Fatalf("HandleWebhook() error = %v", err)
		}
	}
	if ob.n != 1 {
		t.Fatalf("settled events = %d, want 1 for a replayed webhook", ob.n)
	}

	svc.gateways = stubRegistry{stubGateway{name: "momo", res: WebhookResult{PaymentID: "pay-1", Success: false}}}
	if err := svc.HandleWebhook(ctx, WebhookInput{Provider: "momo"}); err != nil {
		t.Fatalf("HandleWebhook() error = %v", err)
	}
	if repo.byID["pay-1"].Status != StatusSucceeded {
		t.Fatalf("status = %s, want succeeded to stay", repo.byID["pay-1"].Status)
	}
}
