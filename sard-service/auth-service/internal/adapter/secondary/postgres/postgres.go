package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

const migrationLockID = 7303_0001

type txKey struct{}

type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type DB struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string, maxConns int32) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() { db.pool.Close() }

func (db *DB) Ping(ctx context.Context) error { return db.pool.Ping(ctx) }

func (db *DB) Migrate(ctx context.Context) error {
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", int64(migrationLockID)); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, schema)
		return err
	})
}

func (db *DB) q(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db.pool
}

func (db *DB) WithLock(ctx context.Context, key string, fn func(ctx context.Context) error) error {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockID(key)); err != nil {
			return err
		}
		return fn(ctx)
	}
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockID(key)); err != nil {
			return err
		}
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

func lockID(key string) int64 {
	h := fnv.New64a()
	h.Write([]byte("auth-service/" + key))
	return int64(h.Sum64())
}

type Challenges struct{ db *DB }

func NewChallenges(db *DB) Challenges { return Challenges{db: db} }

const columns = `id, card_id, account_id, customer_id, user_id, username, masked_pan, amount, currency, merchant, mcc,
	otp_expires_at, attempts, max_attempts, resends, status, failure_reason, expires_at, authenticated_at, valid_until,
	used_at, created_at, updated_at`

func (r Challenges) Create(ctx context.Context, c *domain.Challenge) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO payment_authentications (`+columns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		c.ID, c.CardID, c.AccountID, c.CustomerID, c.UserID, c.Username, c.MaskedPAN, c.Amount, c.Currency, c.Merchant, c.MCC,
		c.OTPExpiresAt, c.Attempts, c.MaxAttempts, c.Resends, string(c.Status), c.FailureReason, c.ExpiresAt, c.AuthenticatedAt,
		c.ValidUntil, c.UsedAt, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r Challenges) Update(ctx context.Context, c *domain.Challenge) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE payment_authentications SET
		attempts = $2, status = $3, failure_reason = $4, authenticated_at = $5, valid_until = $6,
		used_at = $7, updated_at = $8, otp_expires_at = $9, resends = $10 WHERE id = $1`,
		c.ID, c.Attempts, string(c.Status), c.FailureReason, c.AuthenticatedAt, c.ValidUntil, c.UsedAt, c.UpdatedAt,
		c.OTPExpiresAt, c.Resends)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r Challenges) Get(ctx context.Context, id string) (*domain.Challenge, error) {
	return scan(r.db.q(ctx).QueryRow(ctx, `SELECT `+columns+` FROM payment_authentications WHERE id = $1`, id))
}

func (r Challenges) ListPendingByUser(ctx context.Context, userID int64, now time.Time) ([]*domain.Challenge, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+columns+` FROM payment_authentications
		WHERE user_id = $1 AND status = 'PENDING' AND expires_at > $2 ORDER BY created_at DESC LIMIT 50`, userID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Challenge{}
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scan(row pgx.Row) (*domain.Challenge, error) {
	var (
		c      domain.Challenge
		status string
	)
	err := row.Scan(&c.ID, &c.CardID, &c.AccountID, &c.CustomerID, &c.UserID, &c.Username, &c.MaskedPAN, &c.Amount, &c.Currency,
		&c.Merchant, &c.MCC, &c.OTPExpiresAt, &c.Attempts, &c.MaxAttempts, &c.Resends, &status, &c.FailureReason, &c.ExpiresAt,
		&c.AuthenticatedAt, &c.ValidUntil, &c.UsedAt, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.Status = domain.Status(status)
	c.ExpiresAt, c.CreatedAt, c.UpdatedAt = c.ExpiresAt.UTC(), c.CreatedAt.UTC(), c.UpdatedAt.UTC()
	c.OTPExpiresAt = c.OTPExpiresAt.UTC()
	for _, p := range []**time.Time{&c.AuthenticatedAt, &c.ValidUntil, &c.UsedAt} {
		if *p != nil {
			t := (*p).UTC()
			*p = &t
		}
	}
	return &c, nil
}
