package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type txKey struct{}

func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

func forUpdate() clause.Locking { return clause.Locking{Strength: "UPDATE"} }

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}

const (
	mysqlDuplicateEntry = 1062
	mysqlDeadlock       = 1213
	mysqlLockWaitTimout = 1205
)

func mapWriteErr(err error) error {
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == mysqlDuplicateEntry {
		return port.ErrAlreadyExists
	}
	return err
}

func retryable(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && (me.Number == mysqlDeadlock || me.Number == mysqlLockWaitTimout)
}

type txManager struct{ db *gorm.DB }

func NewTxManager(db *gorm.DB) *txManager { return &txManager{db: db} }

func (m *txManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	const attempts = 4
	var err error
	for i := 0; i < attempts; i++ {
		err = m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return fn(context.WithValue(ctx, txKey{}, tx))
		})
		if err == nil || !retryable(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * 20 * time.Millisecond):
		}
	}
	return err
}
