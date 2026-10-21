package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
)

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type WalletRepository interface {
	Create(ctx context.Context, w *model.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Wallet, error)
	GetByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency model.Currency) (*model.Wallet, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Wallet, error)
	Update(ctx context.Context, w *model.Wallet) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
	Update(ctx context.Context, tx *model.Transaction) error
	List(ctx context.Context, filter model.TransactionFilter) ([]*model.Transaction, int64, error)
	ListUnbatchedCaptures(ctx context.Context, merchantID uuid.UUID) ([]*model.Transaction, error)
	ListByBatch(ctx context.Context, batchID uuid.UUID) ([]*model.Transaction, error)
	AssignBatch(ctx context.Context, ids []uuid.UUID, status model.TransactionStatus, batchID uuid.UUID) error
}

type LedgerRepository interface {
	Append(ctx context.Context, entries ...*model.LedgerEntry) error
	List(ctx context.Context, walletID uuid.UUID, from, to *time.Time, page, pageSize int) ([]*model.LedgerEntry, int64, error)
	SumDebitsSince(ctx context.Context, walletID uuid.UUID, since time.Time) (decimal.Decimal, error)
	Summarize(ctx context.Context, walletID uuid.UUID, from, to time.Time) (model.LedgerSummary, error)
}

type CardRepository interface {
	Create(ctx context.Context, card *model.Card) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Card, error)
	ListByWallet(ctx context.Context, walletID uuid.UUID) ([]*model.Card, error)
	Update(ctx context.Context, card *model.Card) error
}

type BankRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Bank, error)
}

type MerchantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error)
}

type BankAccountRepository interface {
	Create(ctx context.Context, a *model.BankAccount) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.BankAccount, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*model.BankAccount, error)
	Update(ctx context.Context, a *model.BankAccount) error
	ClearDefault(ctx context.Context, userID uuid.UUID) error
}

type PaymentRequestRepository interface {
	Create(ctx context.Context, p *model.PaymentRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.PaymentRequest, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.PaymentRequest, error)
	Update(ctx context.Context, p *model.PaymentRequest) error
	List(ctx context.Context, filter model.PaymentRequestFilter) ([]*model.PaymentRequest, int64, error)
}

type SettlementBatchRepository interface {
	Create(ctx context.Context, b *model.SettlementBatch) error
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.SettlementBatch, error)
	Update(ctx context.Context, b *model.SettlementBatch) error
}

type ClearingRepository interface {
	CreateBulk(ctx context.Context, records []*model.ClearingRecord) error
	ListByBatch(ctx context.Context, batchID uuid.UUID) ([]*model.ClearingRecord, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type PinHasher interface {
	Hash(pin string) (string, error)
	Compare(hash, pin string) bool
}

type BankTransfer struct {
	Reference uuid.UUID
	Account   *model.BankAccount
	Amount    decimal.Decimal
	Currency  model.Currency
}

type BankGateway interface {
	Collect(ctx context.Context, t BankTransfer) (string, error)
	Payout(ctx context.Context, t BankTransfer) (string, error)
}
