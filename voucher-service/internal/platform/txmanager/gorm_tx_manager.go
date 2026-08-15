package txmanager

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

type gormTxManager struct {
	db *gorm.DB
}

func NewGormTxManager(db *gorm.DB) TxManager {
	return &gormTxManager{db: db}
}

func (m *gormTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, ctxKey{}, tx))
	})
}

func DBFromContext(ctx context.Context, base *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok {
		return tx
	}
	return base.WithContext(ctx)
}
