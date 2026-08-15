package wallet

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/wallet"
)

type WalletService interface {
	GetOrCreateWallet(ctx context.Context, ownerType wallet.OwnerType, ownerID, currency string) (*wallet.Wallet, error)
	GetBalance(ctx context.Context, ownerType wallet.OwnerType, ownerID string) (shared.Money, error)
	Credit(ctx context.Context, ownerType wallet.OwnerType, ownerID string, amount shared.Money, refType, refID, idempotencyKey string) error
	Debit(ctx context.Context, ownerType wallet.OwnerType, ownerID string, amount shared.Money, refType, refID, idempotencyKey string) error
}

type WalletRepository interface {
	FindByOwner(ctx context.Context, ownerType wallet.OwnerType, ownerID string) (*wallet.Wallet, error)
	FindByOwnerForUpdate(ctx context.Context, ownerType wallet.OwnerType, ownerID string) (*wallet.Wallet, error)
	Create(ctx context.Context, w *wallet.Wallet) error
	Save(ctx context.Context, w *wallet.Wallet) error
}

type LedgerEntry struct {
	WalletID       shared.WalletID
	Type           string // "credit" | "debit"
	Amount         shared.Money
	BalanceAfter   shared.Money
	RefType        string
	RefID          string
	IdempotencyKey string
}

type LedgerRepository interface {
	Append(ctx context.Context, entry LedgerEntry) error
}
