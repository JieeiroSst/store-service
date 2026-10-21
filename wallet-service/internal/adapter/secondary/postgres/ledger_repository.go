package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const ledgerColumns = `id, wallet_id, transaction_id, direction, amount, balance_after, created_at`

type ledgerRepository struct{ base }

func NewLedgerRepository(db *sqlx.DB) port.LedgerRepository {
	return &ledgerRepository{base{db: db}}
}

func (r *ledgerRepository) Append(ctx context.Context, entries ...*model.LedgerEntry) error {
	for _, e := range entries {
		if _, err := r.named(ctx, `INSERT INTO ledger_entries (`+ledgerColumns+`)
			VALUES (:id, :wallet_id, :transaction_id, :direction, :amount, :balance_after, :created_at)`, e); err != nil {
			return err
		}
	}
	return nil
}

func (r *ledgerRepository) List(ctx context.Context, walletID uuid.UUID, from, to *time.Time, page, pageSize int) ([]*model.LedgerEntry, int64, error) {
	where := []string{"wallet_id = $1"}
	args := []any{walletID}
	if from != nil {
		args = append(args, *from)
		where = append(where, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if to != nil {
		args = append(args, *to)
		where = append(where, fmt.Sprintf("created_at < $%d", len(args)))
	}
	cond := strings.Join(where, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE `+cond, args...)
	if err != nil {
		return nil, 0, err
	}
	entries := make([]*model.LedgerEntry, 0)
	query := fmt.Sprintf(`SELECT %s FROM ledger_entries WHERE %s ORDER BY seq DESC LIMIT $%d OFFSET $%d`,
		ledgerColumns, cond, len(args)+1, len(args)+2)
	args = append(args, pageSize, (page-1)*pageSize)
	if err := sqlx.SelectContext(ctx, r.q(ctx), &entries, query, args...); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

func (r *ledgerRepository) SumDebitsSince(ctx context.Context, walletID uuid.UUID, since time.Time) (decimal.Decimal, error) {
	var sum decimal.Decimal
	err := sqlx.GetContext(ctx, r.q(ctx), &sum, `SELECT COALESCE(SUM(amount), 0) FROM ledger_entries
		WHERE wallet_id = $1 AND direction = 'DEBIT' AND created_at >= $2`, walletID, since)
	return sum, err
}

func (r *ledgerRepository) Summarize(ctx context.Context, walletID uuid.UUID, from, to time.Time) (model.LedgerSummary, error) {
	var row struct {
		Opening     decimal.Decimal `db:"opening_balance"`
		TotalCredit decimal.Decimal `db:"total_credit"`
		TotalDebit  decimal.Decimal `db:"total_debit"`
		CreditCount int64           `db:"credit_count"`
		DebitCount  int64           `db:"debit_count"`
	}
	err := sqlx.GetContext(ctx, r.q(ctx), &row, `SELECT
		COALESCE((SELECT balance_after FROM ledger_entries
		          WHERE wallet_id = $1 AND created_at < $2 ORDER BY seq DESC LIMIT 1), 0) AS opening_balance,
		COALESCE(SUM(amount) FILTER (WHERE direction = 'CREDIT'), 0) AS total_credit,
		COALESCE(SUM(amount) FILTER (WHERE direction = 'DEBIT'), 0) AS total_debit,
		COUNT(*) FILTER (WHERE direction = 'CREDIT') AS credit_count,
		COUNT(*) FILTER (WHERE direction = 'DEBIT') AS debit_count
		FROM ledger_entries
		WHERE wallet_id = $1 AND created_at >= $2 AND created_at < $3`, walletID, from, to)
	if err != nil {
		return model.LedgerSummary{}, err
	}
	return model.LedgerSummary{
		OpeningBalance: row.Opening,
		ClosingBalance: row.Opening.Add(row.TotalCredit).Sub(row.TotalDebit),
		TotalCredit:    row.TotalCredit,
		TotalDebit:     row.TotalDebit,
		CreditCount:    row.CreditCount,
		DebitCount:     row.DebitCount,
	}, nil
}
