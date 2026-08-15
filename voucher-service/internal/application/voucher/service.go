package voucher

import (
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/JIeeiroSst/voucher-service/internal/platform/lock"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"go.uber.org/zap"
)

const aggregateType = "voucher"
const outboxTopic = "voucher.events"

type Service struct {
	repo           VoucherRepository
	registry       ProviderRegistry
	merchantLookup MerchantLookup
	txManager      txmanager.TxManager
	locker         lock.Locker
	idemp          idempotency.Store
	outboxP        outbox.Outbox
	clock          shared.Clock
	log            *zap.Logger
}

func NewService(
	repo VoucherRepository,
	registry ProviderRegistry,
	merchantLookup MerchantLookup,
	txManager txmanager.TxManager,
	locker lock.Locker,
	idemp idempotency.Store,
	outboxP outbox.Outbox,
	clock shared.Clock,
	log *zap.Logger,
) VoucherService {
	return &Service{
		repo:           repo,
		registry:       registry,
		merchantLookup: merchantLookup,
		txManager:      txManager,
		locker:         locker,
		idemp:          idemp,
		outboxP:        outboxP,
		clock:          clock,
		log:            log,
	}
}
