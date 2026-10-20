package repository

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
)

type PointLedgerRepository interface {
	Credit(ctx context.Context, tx *entity.PointTransaction) (bool, error)
	GetAccount(ctx context.Context, userID string) (*entity.PointAccount, error)
	ListTransactions(ctx context.Context, userID string, limit, offset int) ([]entity.PointTransaction, int64, error)
}
