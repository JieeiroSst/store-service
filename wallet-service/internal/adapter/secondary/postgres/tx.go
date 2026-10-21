package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type txKey struct{}

type TxManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) port.TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return fn(ctx)
	}
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type base struct {
	db *sqlx.DB
}

func (b base) q(ctx context.Context) sqlx.ExtContext {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return b.db
}

func (b base) get(ctx context.Context, dest any, notFound error, query string, args ...any) error {
	err := sqlx.GetContext(ctx, b.q(ctx), dest, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}
	return err
}

func (b base) count(ctx context.Context, query string, args ...any) (int64, error) {
	var n int64
	err := sqlx.GetContext(ctx, b.q(ctx), &n, query, args...)
	return n, err
}

func (b base) named(ctx context.Context, query string, arg any) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, b.q(ctx), query, arg)
}

func uniqueViolation(err error, constraint string) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
		return false
	}
	return constraint == "" || pqErr.Constraint == constraint
}

func expectOne(res sql.Result, notFound error) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return notFound
	}
	return nil
}
