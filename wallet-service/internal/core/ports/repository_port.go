package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *domain.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	UpdateWithVersion(ctx context.Context, wallet *domain.Wallet) error
	List(ctx context.Context, page, pageSize int) ([]*domain.Wallet, int64, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error)
	Update(ctx context.Context, tx *domain.Transaction) error
	ListByWalletID(ctx context.Context, walletID uuid.UUID, page, pageSize int) ([]*domain.Transaction, int64, error)
	ListCapturedByMerchant(ctx context.Context, merchantID uuid.UUID, since time.Time) ([]*domain.Transaction, error)
	UpdateBatch(ctx context.Context, txIDs []uuid.UUID, status domain.TransactionStatus, batchID uuid.UUID) error
}

type CardRepository interface {
	Create(ctx context.Context, card *domain.Card) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Card, error)
	GetByWalletID(ctx context.Context, walletID uuid.UUID) ([]*domain.Card, error)
	Update(ctx context.Context, card *domain.Card) error
}

type SettlementBatchRepository interface {
	Create(ctx context.Context, batch *domain.SettlementBatch) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SettlementBatch, error)
	Update(ctx context.Context, batch *domain.SettlementBatch) error
}

type ClearingRepository interface {
	CreateBulk(ctx context.Context, records []*domain.ClearingRecord) error
	GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.ClearingRecord, error)
}

type BankRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bank, error)
}

type MerchantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
}

type FeeCalculator interface {
	Calculate(amount decimal.Decimal, network domain.CardNetwork) decimal.Decimal
}
