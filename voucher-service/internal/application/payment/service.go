package payment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"go.uber.org/zap"
)

const settledEventType = "payment.settled"
const settledTopic = "payment.events"

type PaymentSettledEvent struct {
	OrderID    string `json:"order_id"`
	PaymentRef string `json:"payment_ref"`
}

type Service struct {
	repo      PaymentRepository
	gateways  PaymentGatewayRegistry
	txManager txmanager.TxManager
	outboxP   outbox.Outbox
	log       *zap.Logger
}

func NewService(repo PaymentRepository, gateways PaymentGatewayRegistry, txManager txmanager.TxManager, outboxP outbox.Outbox, log *zap.Logger) PaymentService {
	return &Service{repo: repo, gateways: gateways, txManager: txManager, outboxP: outboxP, log: log}
}

func (s *Service) InitiatePayment(ctx context.Context, in InitiatePaymentInput) (*InitiatePaymentOutput, error) {
	p := &Payment{
		ID:       shared.NewVoucherID().String(),
		OrderID:  in.OrderID,
		Provider: in.Provider,
		Amount:   in.Amount,
		Status:   StatusPending,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	gateway, err := s.gateways.Resolve(in.Provider)
	if err != nil {
		return nil, err
	}
	redirectURL, err := gateway.InitPayment(ctx, p.ID, in.Amount, in.ReturnURL)
	if err != nil {
		return nil, err
	}

	return &InitiatePaymentOutput{PaymentID: p.ID, RedirectURL: redirectURL}, nil
}

func (s *Service) HandleWebhook(ctx context.Context, in WebhookInput) error {
	gateway, err := s.gateways.Resolve(in.Provider)
	if err != nil {
		return err
	}
	paymentID, success, err := gateway.VerifyWebhookSignature(in.RawBody, in.Signature)
	if err != nil {
		return err
	}
	if !success {
		p, err := s.repo.FindByID(ctx, paymentID)
		if err != nil {
			return err
		}
		p.Status = StatusFailed
		return s.repo.Save(ctx, p)
	}

	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		p, err := s.repo.FindByID(ctx, paymentID)
		if err != nil {
			return err
		}
		p.Status = StatusSucceeded
		p.ProviderTxnRef = paymentID
		if err := s.repo.Save(ctx, p); err != nil {
			return err
		}

		payload, err := json.Marshal(PaymentSettledEvent{OrderID: p.OrderID.String(), PaymentRef: p.ProviderTxnRef})
		if err != nil {
			return err
		}
		return s.outboxP.Enqueue(ctx, outbox.Event{
			AggregateType: "payment",
			AggregateID:   p.ID,
			EventType:     settledEventType,
			Payload:       payload,
			Topic:         settledTopic,
			OccurredAt:    time.Now().UTC(),
		})
	})
}

func (s *Service) Refund(ctx context.Context, paymentID string, amount shared.Money) error {
	p, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}
	gateway, err := s.gateways.Resolve(p.Provider)
	if err != nil {
		return err
	}
	if err := gateway.Refund(ctx, p.ProviderTxnRef, amount); err != nil {
		return err
	}
	p.Status = StatusRefunded
	return s.repo.Save(ctx, p)
}

func (s *Service) GetPayment(ctx context.Context, id string) (*Payment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListSettledSince(ctx context.Context, since time.Time) ([]*Payment, error) {
	return s.repo.ListSettledSince(ctx, since)
}
