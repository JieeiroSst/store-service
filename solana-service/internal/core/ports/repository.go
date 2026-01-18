package ports

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
)

type TransactionRepository interface {
	Save(ctx context.Context, tx *domain.Transaction) error
	FindBySignature(ctx context.Context, signature string) (*domain.Transaction, error)
	FindByAddress(ctx context.Context, address string) ([]*domain.Transaction, error)
}

type AccountRepository interface {
	Save(ctx context.Context, account *domain.Account) error
	FindByPublicKey(ctx context.Context, pubkey string) (*domain.Account, error)
}

type BridgeRepository interface {
	Save(ctx context.Context, transfer *domain.BridgeTransfer) error
	FindByID(ctx context.Context, id string) (*domain.BridgeTransfer, error)
	FindBySolanaAddress(ctx context.Context, address string) ([]*domain.BridgeTransfer, error)
	FindByCircleWallet(ctx context.Context, walletID string) ([]*domain.BridgeTransfer, error)
	Update(ctx context.Context, transfer *domain.BridgeTransfer) error
}