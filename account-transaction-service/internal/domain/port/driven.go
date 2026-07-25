package port

import (
	"context"

	"github.com/Jieeirosst/account-transaction-service/internal/domain/model"
	"github.com/Jieeirosst/account-transaction-service/pkg/pagination"
)

// AccountRepository is the secondary port for account persistence.
type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	GetByID(ctx context.Context, id string) (*model.Account, error)
	List(ctx context.Context, p pagination.Page) ([]model.Account, error)
}

// TransactionRepository is the secondary port for transaction persistence.
type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	GetByID(ctx context.Context, id string) (*model.Transaction, error)
	List(ctx context.Context, p pagination.Page, txType model.TransactionType) ([]model.Transaction, error)
	ListByAccount(ctx context.Context, accountID string, p pagination.Page) ([]model.Transaction, error)
}
