package order

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"go.uber.org/zap"
)

const aggregateType = "order"
const outboxTopic = "order.events"

type Service struct {
	repo             OrderRepository
	voucherIssuer    VoucherIssuer
	paymentInitiator PaymentInitiator
	walletDebiter    WalletDebiter
	txManager        txmanager.TxManager
	outboxP          outbox.Outbox
	clock            shared.Clock
	log              *zap.Logger
}

func NewService(
	repo OrderRepository,
	voucherIssuer VoucherIssuer,
	paymentInitiator PaymentInitiator,
	walletDebiter WalletDebiter,
	txManager txmanager.TxManager,
	outboxP outbox.Outbox,
	clock shared.Clock,
	log *zap.Logger,
) OrderService {
	return &Service{
		repo:             repo,
		voucherIssuer:    voucherIssuer,
		paymentInitiator: paymentInitiator,
		walletDebiter:    walletDebiter,
		txManager:        txManager,
		outboxP:          outboxP,
		clock:            clock,
		log:              log,
	}
}

func (s *Service) enqueueEvents(ctx context.Context, events []shared.DomainEvent) error {
	for _, evt := range events {
		outboxEvt, err := outbox.NewEventFromDomain(aggregateType, outboxTopic, evt)
		if err != nil {
			return err
		}
		if err := s.outboxP.Enqueue(ctx, outboxEvt); err != nil {
			return err
		}
	}
	return nil
}
