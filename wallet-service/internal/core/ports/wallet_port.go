package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
)

type WalletUseCase interface {
	CreateWallet(ctx context.Context, userID uuid.UUID, currency domain.Currency) (*domain.Wallet, error)
	GetWallet(ctx context.Context, walletID uuid.UUID) (*domain.Wallet, error)
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	Credit(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, description string) (*domain.Transaction, error)
	FreezeWallet(ctx context.Context, walletID uuid.UUID) error
	UnfreezeWallet(ctx context.Context, walletID uuid.UUID) error
}

type TransactionUseCase interface {
	Authorize(ctx context.Context, req AuthorizeRequest) (*domain.Transaction, error)
	Capture(ctx context.Context, transactionID uuid.UUID) (*domain.Transaction, error)
	Void(ctx context.Context, transactionID uuid.UUID) (*domain.Transaction, error)
	CreateSettlementBatch(ctx context.Context, merchantID uuid.UUID) (*domain.SettlementBatch, error)
	ProcessClearing(ctx context.Context, batchID uuid.UUID) ([]*domain.ClearingRecord, error)
	ProcessSettlement(ctx context.Context, batchID uuid.UUID) error

	GetTransaction(ctx context.Context, transactionID uuid.UUID) (*domain.Transaction, error)
	ListTransactionsByWallet(ctx context.Context, walletID uuid.UUID, page, pageSize int) ([]*domain.Transaction, int64, error)
}

type CardUseCase interface {
	IssueCard(ctx context.Context, walletID uuid.UUID, network domain.CardNetwork, holderName string) (*domain.Card, error)
	GetCard(ctx context.Context, cardID uuid.UUID) (*domain.Card, error)
	BlockCard(ctx context.Context, cardID uuid.UUID) error
	ListCardsByWallet(ctx context.Context, walletID uuid.UUID) ([]*domain.Card, error)
}

type AuthorizeRequest struct {
	IdempotencyKey string
	CardID         uuid.UUID
	MerchantID     uuid.UUID
	Amount         decimal.Decimal
	Currency       domain.Currency
	Description    string
	Metadata       map[string]string
}
