package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
)

type WalletUsecase interface {
	CreateWallet(ctx context.Context, userID, currency string) (*model.Wallet, error)
	GetWallet(ctx context.Context, walletID string) (*model.Wallet, error)
	GetWalletByUser(ctx context.Context, userID string) (*model.Wallet, error)

	FreezeWallet(ctx context.Context, walletID, reason string) (*model.Wallet, error)
	UnfreezeWallet(ctx context.Context, walletID string) (*model.Wallet, error)
	CloseWallet(ctx context.Context, walletID string) (*model.Wallet, error)
	SetLimits(ctx context.Context, walletID string, dailyLimit, perTransactionLimit int64) (*model.Wallet, error)
}

type TransactionUsecase interface {
	Deposit(ctx context.Context, walletID string, amount int64, referenceID, description string) (*model.Transaction, error)
	Withdraw(ctx context.Context, walletID string, amount int64, referenceID, description string) (*model.Transaction, error)
	Transfer(ctx context.Context, senderWalletID, receiverWalletID string, amount int64, referenceID, description string) (*model.Transfer, error)
	GetTransaction(ctx context.Context, transactionID string) (*model.Transaction, error)
	ListTransactions(ctx context.Context, walletID string, limit, offset int) ([]model.Transaction, error)
	Statement(ctx context.Context, walletID string, from, to time.Time) ([]model.Transaction, error)

	GetTransfer(ctx context.Context, transferID string) (*model.Transfer, error)
	Reverse(ctx context.Context, transactionID, reason string) (*model.Transaction, error)
	ReverseTransfer(ctx context.Context, transferID, reason string) (*model.Transfer, error)
}

type PaymentMethodUsecase interface {
	AddPaymentMethod(ctx context.Context, userID string, methodType model.PaymentMethodType, provider, accountNumber string) (*model.PaymentMethod, error)
	ListPaymentMethods(ctx context.Context, userID string) ([]model.PaymentMethod, error)
	RemovePaymentMethod(ctx context.Context, id string) error
	SetDefaultPaymentMethod(ctx context.Context, id string) (*model.PaymentMethod, error)
}

type PocketUsecase interface {
	CreatePocket(ctx context.Context, walletID, name string) (*model.Pocket, error)
	ListPockets(ctx context.Context, walletID string) ([]model.Pocket, error)
	DepositToPocket(ctx context.Context, pocketID string, amount int64) (*model.Pocket, error)
	WithdrawFromPocket(ctx context.Context, pocketID string, amount int64) (*model.Pocket, error)
	ClosePocket(ctx context.Context, pocketID string) (*model.Wallet, error)
}

type PaymentRequestUsecase interface {
	CreatePaymentRequest(ctx context.Context, requesterWalletID string, payerWalletID *string, amount int64, description string, expiresIn time.Duration) (*model.PaymentRequest, error)
	GetPaymentRequest(ctx context.Context, id string) (*model.PaymentRequest, error)
	ListPaymentRequests(ctx context.Context, walletID string) ([]model.PaymentRequest, error)
	Pay(ctx context.Context, id, payerWalletID string) (*model.Transfer, error)
	Cancel(ctx context.Context, id string) error
}
