package repository

import (
	"context"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"gorm.io/gorm"
)

type txRunner struct{ db *gorm.DB }

func NewTxRunner(db *gorm.DB) port.TxRunner { return &txRunner{db: db} }

func (t *txRunner) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}
