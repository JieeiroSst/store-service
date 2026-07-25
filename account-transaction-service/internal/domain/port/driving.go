package port

import (
	"context"

	"github.com/Jieeirosst/account-transaction-service/internal/domain/model"
)

// AccountUsecase is the primary port for account operations.
type AccountUsecase interface {
	CreateAccount(ctx context.Context, firstName, lastName string) (*model.Account, error)
	GetAccount(ctx context.Context, id string) (*model.Account, error)
	ListAccounts(ctx context.Context, pageSize int32, pageToken string) (accounts []model.Account, nextPageToken string, err error)
}

// TransactionUsecase is the primary port for transaction operations.
type TransactionUsecase interface {
	CreateDeposit(ctx context.Context, accountID string, amount float64) (*model.Transaction, error)
	CreateWithdrawal(ctx context.Context, accountID string, amount float64) (*model.Transaction, error)
	CreateTransfer(ctx context.Context, senderID, receiverID string, amount float64) (*model.Transaction, error)
	CreatePayment(ctx context.Context, accountID, serviceName string, amount float64) (*model.Transaction, error)

	GetTransaction(ctx context.Context, id string) (*model.Transaction, error)
	ListTransactions(ctx context.Context, pageSize int32, pageToken string, txType string) (txs []model.Transaction, nextPageToken string, err error)
	GetAccountTransactions(ctx context.Context, accountID string, pageSize int32, pageToken string) (txs []model.Transaction, nextPageToken string, err error)
}
