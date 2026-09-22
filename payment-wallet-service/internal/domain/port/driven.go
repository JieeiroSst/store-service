package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *model.Wallet) error
	GetByID(ctx context.Context, walletID string) (*model.Wallet, error)
	GetByUserID(ctx context.Context, userID string) (*model.Wallet, error)

	Deposit(ctx context.Context, walletID string, txn *model.Transaction) (*model.Wallet, error)
	Withdraw(ctx context.Context, walletID string, txn *model.Transaction) (*model.Wallet, error)
	Transfer(ctx context.Context, senderWalletID, receiverWalletID string, transfer *model.Transfer, outTxn, inTxn *model.Transaction) (sender *model.Wallet, receiver *model.Wallet, err error)

	Reverse(ctx context.Context, walletID string, original *model.Transaction, reversal *model.Transaction) (*model.Wallet, error)
	ReverseTransfer(ctx context.Context, transfer *model.Transfer, reversalToSender, reversalFromReceiver *model.Transaction) (sender *model.Wallet, receiver *model.Wallet, err error)

	UpdateStatus(ctx context.Context, walletID string, status model.WalletStatus, reason *string) (*model.Wallet, error)
	Close(ctx context.Context, walletID string) (*model.Wallet, error)
	UpdateLimits(ctx context.Context, walletID string, dailyLimit, perTransactionLimit int64) (*model.Wallet, error)
}

type TransactionRepository interface {
	GetByID(ctx context.Context, transactionID string) (*model.Transaction, error)
	GetByReferenceID(ctx context.Context, walletID, referenceID string) (*model.Transaction, error)
	ListByWallet(ctx context.Context, walletID string, limit, offset int) ([]model.Transaction, error)
	ListByWalletAndDateRange(ctx context.Context, walletID string, from, to time.Time) ([]model.Transaction, error)
	SumOutgoingSince(ctx context.Context, walletID string, since time.Time) (int64, error)
}

type TransferRepository interface {
	GetByID(ctx context.Context, transferID string) (*model.Transfer, error)
	GetByReferenceID(ctx context.Context, referenceID string) (*model.Transfer, error)
}

type PaymentMethodRepository interface {
	Create(ctx context.Context, method *model.PaymentMethod) error
	GetByID(ctx context.Context, id string) (*model.PaymentMethod, error)
	ListByUser(ctx context.Context, userID string) ([]model.PaymentMethod, error)
	Delete(ctx context.Context, id string) error
	SetDefault(ctx context.Context, id string) (*model.PaymentMethod, error)
}

type PocketRepository interface {
	Create(ctx context.Context, pocket *model.Pocket) error
	GetByID(ctx context.Context, pocketID string) (*model.Pocket, error)
	ListByWallet(ctx context.Context, walletID string) ([]model.Pocket, error)
	MoveToPocket(ctx context.Context, walletID, pocketID string, txn *model.Transaction) (*model.Wallet, *model.Pocket, error)
	MoveFromPocket(ctx context.Context, walletID, pocketID string, txn *model.Transaction) (*model.Wallet, *model.Pocket, error)
	Close(ctx context.Context, walletID, pocketID string) (*model.Wallet, error)
}

type PaymentRequestRepository interface {
	Create(ctx context.Context, request *model.PaymentRequest) error
	GetByID(ctx context.Context, id string) (*model.PaymentRequest, error)
	ListByWallet(ctx context.Context, walletID string) ([]model.PaymentRequest, error)
	UpdateStatus(ctx context.Context, id string, status model.PaymentRequestStatus, transferID *string) error
}
