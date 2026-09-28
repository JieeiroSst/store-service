package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type txKey struct{}

type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) port.TxManager {
	return &txManager{db: db}
}

func (m *txManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

var forUpdate = clause.Locking{Strength: "UPDATE"}

func translate(err error, what string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("%w: %s", port.ErrNotFound, what)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return fmt.Errorf("%w: %s already exists", port.ErrConflict, what)
	default:
		return err
	}
}
