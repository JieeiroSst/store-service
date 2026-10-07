package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

//go:embed schema.sql
var schema string

const migrationLockID = 7301_0001

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
	h.Write([]byte("card-service/" + key))
	return int64(h.Sum64())
}

type Accounts struct{ db *DB }

func NewAccounts(db *DB) Accounts { return Accounts{db: db} }

const accountColumns = `id, customer_id, user_id, program_code, mode, currency, holder_name, status, status_reason,
	credit_limit, balance, held, created_at, updated_at`

func (r Accounts) Create(ctx context.Context, a *domain.Account) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO accounts (`+accountColumns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		a.ID, a.CustomerID, a.UserID, a.ProgramCode, string(a.Mode), a.Currency, a.HolderName, string(a.Status),
		a.StatusReason, a.CreditLimit, a.Balance, a.Held, a.CreatedAt, a.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.Conflict("customer already has an open account in program %s", a.ProgramCode)
	}
	return err
}

func (r Accounts) Update(ctx context.Context, a *domain.Account) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE accounts SET status = $2, status_reason = $3, credit_limit = $4,
		balance = $5, held = $6, updated_at = $7 WHERE id = $1`,
		a.ID, string(a.Status), a.StatusReason, a.CreditLimit, a.Balance, a.Held, a.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r Accounts) Get(ctx context.Context, id string) (*domain.Account, error) {
	return scanAccount(r.db.q(ctx).QueryRow(ctx, `SELECT `+accountColumns+` FROM accounts WHERE id = $1`, id))
}

func (r Accounts) ListByCustomer(ctx context.Context, customerID string) ([]*domain.Account, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+accountColumns+` FROM accounts WHERE customer_id = $1 ORDER BY created_at, id`, customerID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanAccount)
}

func scanAccount(row pgx.Row) (*domain.Account, error) {
	var (
		a            domain.Account
		mode, status string
	)
	err := row.Scan(&a.ID, &a.CustomerID, &a.UserID, &a.ProgramCode, &mode, &a.Currency, &a.HolderName, &status,
		&a.StatusReason, &a.CreditLimit, &a.Balance, &a.Held, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.Mode, a.Status = domain.Mode(mode), domain.AccountStatus(status)
	a.CreatedAt, a.UpdatedAt = a.CreatedAt.UTC(), a.UpdatedAt.UTC()
	return &a, nil
}

type Cards struct{ db *DB }

func NewCards(db *DB) Cards { return Cards{db: db} }

const cardColumns = `id, account_id, customer_id, program_code, card_type, cardholder_name, pan_hash, pan_cipher, bin, last4,
	expiry_month, expiry_year, valid_until, service_code, status, status_reason,
	limit_per_transaction, limit_daily, control_pos, control_contactless, control_ecommerce, control_atm,
	pin_hash, pin_failures, replaces_card_id, replaced_by_card_id, created_at, updated_at, activated_at`

func (r Cards) Create(ctx context.Context, c *domain.Card) error {
	tag, err := r.db.q(ctx).Exec(ctx, `INSERT INTO cards (`+cardColumns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29)
		ON CONFLICT (pan_hash) DO NOTHING`,
		c.ID, c.AccountID, c.CustomerID, c.ProgramCode, string(c.Type), c.CardholderName, c.PANHash, c.PANCipher, c.BIN, c.Last4,
		c.Expiry.Month, c.Expiry.Year, c.ValidUntil, c.ServiceCode, string(c.Status), c.StatusReason,
		c.Limits.PerTransaction, c.Limits.Daily, c.Controls.POS, c.Controls.Contactless, c.Controls.Ecommerce, c.Controls.ATM,
		c.PINHash, c.PINFailures, c.ReplacesCardID, c.ReplacedByCardID, c.CreatedAt, c.UpdatedAt, c.ActivatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDuplicatePAN
	}
	return nil
}

func (r Cards) Update(ctx context.Context, c *domain.Card) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE cards SET
		status = $2, status_reason = $3, limit_per_transaction = $4, limit_daily = $5,
		control_pos = $6, control_contactless = $7, control_ecommerce = $8, control_atm = $9,
		pin_hash = $10, pin_failures = $11, replaced_by_card_id = $12, updated_at = $13, activated_at = $14
		WHERE id = $1`,
		c.ID, string(c.Status), c.StatusReason, c.Limits.PerTransaction, c.Limits.Daily,
		c.Controls.POS, c.Controls.Contactless, c.Controls.Ecommerce, c.Controls.ATM,
		c.PINHash, c.PINFailures, c.ReplacedByCardID, c.UpdatedAt, c.ActivatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r Cards) Get(ctx context.Context, id string) (*domain.Card, error) {
	return scanCard(r.db.q(ctx).QueryRow(ctx, `SELECT `+cardColumns+` FROM cards WHERE id = $1`, id))
}

func (r Cards) GetByPANHash(ctx context.Context, hash string) (*domain.Card, error) {
	return scanCard(r.db.q(ctx).QueryRow(ctx, `SELECT `+cardColumns+` FROM cards WHERE pan_hash = $1`, hash))
}

func (r Cards) ListByAccount(ctx context.Context, accountID string) ([]*domain.Card, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+cardColumns+` FROM cards WHERE account_id = $1 ORDER BY created_at, id`, accountID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanCard)
}

func scanCard(row pgx.Row) (*domain.Card, error) {
	var (
		c              domain.Card
		cardType, stat string
	)
	err := row.Scan(
		&c.ID, &c.AccountID, &c.CustomerID, &c.ProgramCode, &cardType, &c.CardholderName, &c.PANHash, &c.PANCipher, &c.BIN, &c.Last4,
		&c.Expiry.Month, &c.Expiry.Year, &c.ValidUntil, &c.ServiceCode, &stat, &c.StatusReason,
		&c.Limits.PerTransaction, &c.Limits.Daily, &c.Controls.POS, &c.Controls.Contactless, &c.Controls.Ecommerce, &c.Controls.ATM,
		&c.PINHash, &c.PINFailures, &c.ReplacesCardID, &c.ReplacedByCardID, &c.CreatedAt, &c.UpdatedAt, &c.ActivatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.Type, c.Status = domain.CardType(cardType), domain.Status(stat)
	c.CreatedAt, c.UpdatedAt = c.CreatedAt.UTC(), c.UpdatedAt.UTC()
	for _, p := range []**time.Time{&c.ActivatedAt, &c.ValidUntil} {
		if *p != nil {
			t := (*p).UTC()
			*p = &t
		}
	}
	return &c, nil
}

type Authorizations struct{ db *DB }

func NewAuthorizations(db *DB) Authorizations { return Authorizations{db: db} }

const authColumns = `id, account_id, card_id, processing_code, status, amount, confirmed_amount, currency, channel,
	merchant, mcc, response_code, reason, auth_code, authentication_id, created_at, updated_at`

func (r Authorizations) Create(ctx context.Context, a *domain.Authorization) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO authorizations (`+authColumns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		a.ID, a.AccountID, a.CardID, string(a.ProcessingCode), string(a.Status), a.Amount, a.ConfirmedAmount, a.Currency,
		string(a.Channel), a.Merchant, a.MCC, string(a.Code), a.Reason, a.AuthCode, a.AuthenticationID, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r Authorizations) Update(ctx context.Context, a *domain.Authorization) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE authorizations SET status = $2, confirmed_amount = $3, updated_at = $4 WHERE id = $1`,
		a.ID, string(a.Status), a.ConfirmedAmount, a.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r Authorizations) Get(ctx context.Context, id string) (*domain.Authorization, error) {
	return scanAuth(r.db.q(ctx).QueryRow(ctx, `SELECT `+authColumns+` FROM authorizations WHERE id = $1`, id))
}

func (r Authorizations) ListByCard(ctx context.Context, cardID string, limit int) ([]*domain.Authorization, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+authColumns+` FROM authorizations
		WHERE card_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2`, cardID, limit)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanAuth)
}

func (r Authorizations) SumDebits(ctx context.Context, cardID string, since time.Time) (int64, error) {
	var sum int64
	err := r.db.q(ctx).QueryRow(ctx, `SELECT COALESCE(SUM(CASE WHEN status = 'CONFIRMED' THEN confirmed_amount ELSE amount END), 0)::bigint
		FROM authorizations
		WHERE card_id = $1 AND status IN ('AUTHORIZED', 'CONFIRMED') AND processing_code IN ('00', '01') AND created_at >= $2`,
		cardID, since).Scan(&sum)
	return sum, err
}

func scanAuth(row pgx.Row) (*domain.Authorization, error) {
	var (
		a                         domain.Authorization
		pc, status, channel, code string
	)
	err := row.Scan(&a.ID, &a.AccountID, &a.CardID, &pc, &status, &a.Amount, &a.ConfirmedAmount, &a.Currency, &channel,
		&a.Merchant, &a.MCC, &code, &a.Reason, &a.AuthCode, &a.AuthenticationID, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.ProcessingCode, a.Status = domain.ProcessingCode(pc), domain.AuthorizationStatus(status)
	a.Channel, a.Code = domain.Channel(channel), domain.ResponseCode(code)
	a.CreatedAt, a.UpdatedAt = a.CreatedAt.UTC(), a.UpdatedAt.UTC()
	return &a, nil
}

type Transactions struct{ db *DB }

func NewTransactions(db *DB) Transactions { return Transactions{db: db} }

const txnColumns = `id, account_id, card_id, authorization_id, type, processing_code, amount, balance_after, description, created_at`

func (r Transactions) Create(ctx context.Context, t *domain.Transaction) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO transactions (`+txnColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.AccountID, t.CardID, t.AuthorizationID, string(t.Type), string(t.ProcessingCode), t.Amount, t.BalanceAfter,
		t.Description, t.CreatedAt)
	return err
}

func (r Transactions) ListByAccount(ctx context.Context, accountID string, limit int) ([]*domain.Transaction, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+txnColumns+` FROM transactions
		WHERE account_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, err
	}
	return collect(rows, func(row pgx.Row) (*domain.Transaction, error) {
		var (
			t       domain.Transaction
			typ, pc string
		)
		if err := row.Scan(&t.ID, &t.AccountID, &t.CardID, &t.AuthorizationID, &typ, &pc, &t.Amount, &t.BalanceAfter,
			&t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Type, t.ProcessingCode, t.CreatedAt = domain.TransactionType(typ), domain.ProcessingCode(pc), t.CreatedAt.UTC()
		return &t, nil
	})
}

func collect[T any](rows pgx.Rows, scan func(pgx.Row) (*T, error)) ([]*T, error) {
	defer rows.Close()
	out := []*T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
