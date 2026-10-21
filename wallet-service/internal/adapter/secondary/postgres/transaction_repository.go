package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const transactionColumns = `id, COALESCE(idempotency_key, '') AS idempotency_key, wallet_id,
	counterparty_wallet_id, merchant_id, acquirer_bank_id, issuer_bank_id, card_id, bank_account_id,
	reference_id, COALESCE(card_network, '') AS card_network, amount, currency, fee, refunded_amount,
	type, status, COALESCE(description, '') AS description,
	COALESCE(authorization_code, '') AS authorization_code, COALESCE(external_ref, '') AS external_ref,
	COALESCE(failure_reason, '') AS failure_reason, batch_id, metadata, created_at, updated_at`

type transactionRepository struct{ base }

func NewTransactionRepository(db *sqlx.DB) port.TransactionRepository {
	return &transactionRepository{base{db: db}}
}

func (r *transactionRepository) Create(ctx context.Context, tx *model.Transaction) error {
	_, err := r.named(ctx, `INSERT INTO transactions
		(id, idempotency_key, wallet_id, counterparty_wallet_id, merchant_id, acquirer_bank_id,
		 issuer_bank_id, card_id, bank_account_id, reference_id, card_network, amount, currency, fee,
		 refunded_amount, type, status, description, authorization_code, external_ref, failure_reason,
		 batch_id, metadata, created_at, updated_at)
		VALUES
		(:id, NULLIF(:idempotency_key, ''), :wallet_id, :counterparty_wallet_id, :merchant_id,
		 :acquirer_bank_id, :issuer_bank_id, :card_id, :bank_account_id, :reference_id,
		 NULLIF(:card_network, ''), :amount, :currency, :fee, :refunded_amount, :type, :status,
		 NULLIF(:description, ''), NULLIF(:authorization_code, ''), NULLIF(:external_ref, ''),
		 NULLIF(:failure_reason, ''), :batch_id, :metadata, :created_at, :updated_at)`, tx)
	if uniqueViolation(err, "transactions_idempotency_key_key") {
		return model.ErrDuplicateTransaction
	}
	return err
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	var tx model.Transaction
	if err := r.get(ctx, &tx, model.ErrTransactionNotFound, `SELECT `+transactionColumns+` FROM transactions WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	var tx model.Transaction
	if err := r.get(ctx, &tx, model.ErrTransactionNotFound, `SELECT `+transactionColumns+` FROM transactions WHERE id = $1 FOR UPDATE`, id); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	var tx model.Transaction
	if err := r.get(ctx, &tx, model.ErrTransactionNotFound, `SELECT `+transactionColumns+` FROM transactions WHERE idempotency_key = $1`, key); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) Update(ctx context.Context, tx *model.Transaction) error {
	res, err := r.named(ctx, `UPDATE transactions
		SET status = :status, type = :type, refunded_amount = :refunded_amount,
		    authorization_code = NULLIF(:authorization_code, ''), external_ref = NULLIF(:external_ref, ''),
		    failure_reason = NULLIF(:failure_reason, ''), reference_id = :reference_id,
		    batch_id = :batch_id, updated_at = :updated_at
		WHERE id = :id`, tx)
	if err != nil {
		return err
	}
	return expectOne(res, model.ErrTransactionNotFound)
}

func (r *transactionRepository) List(ctx context.Context, f model.TransactionFilter) ([]*model.Transaction, int64, error) {
	where := []string{"wallet_id = $1"}
	args := []any{f.WalletID}
	add := func(clause string, v any) {
		args = append(args, v)
		where = append(where, fmt.Sprintf(clause, len(args)))
	}
	if len(f.Types) > 0 {
		types := make([]string, len(f.Types))
		for i, t := range f.Types {
			types[i] = string(t)
		}
		add("type = ANY($%d)", pq.Array(types))
	}
	if len(f.Statuses) > 0 {
		statuses := make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			statuses[i] = string(s)
		}
		add("status = ANY($%d)", pq.Array(statuses))
	}
	if f.From != nil {
		add("created_at >= $%d", *f.From)
	}
	if f.To != nil {
		add("created_at < $%d", *f.To)
	}
	cond := strings.Join(where, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) FROM transactions WHERE `+cond, args...)
	if err != nil {
		return nil, 0, err
	}
	txs := make([]*model.Transaction, 0)
	query := fmt.Sprintf(`SELECT %s FROM transactions WHERE %s ORDER BY created_at DESC, id LIMIT $%d OFFSET $%d`,
		transactionColumns, cond, len(args)+1, len(args)+2)
	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	if err := sqlx.SelectContext(ctx, r.q(ctx), &txs, query, args...); err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}

func (r *transactionRepository) ListUnbatchedCaptures(ctx context.Context, merchantID uuid.UUID) ([]*model.Transaction, error) {
	txs := make([]*model.Transaction, 0)
	err := sqlx.SelectContext(ctx, r.q(ctx), &txs, `SELECT `+transactionColumns+` FROM transactions
		WHERE merchant_id = $1 AND status = 'CAPTURED' AND card_id IS NOT NULL AND batch_id IS NULL
		ORDER BY created_at ASC
		FOR UPDATE`, merchantID)
	return txs, err
}

func (r *transactionRepository) ListByBatch(ctx context.Context, batchID uuid.UUID) ([]*model.Transaction, error) {
	txs := make([]*model.Transaction, 0)
	err := sqlx.SelectContext(ctx, r.q(ctx), &txs, `SELECT `+transactionColumns+` FROM transactions
		WHERE batch_id = $1 ORDER BY created_at ASC`, batchID)
	return txs, err
}

func (r *transactionRepository) AssignBatch(ctx context.Context, ids []uuid.UUID, status model.TransactionStatus, batchID uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	raw := make([]string, len(ids))
	for i, id := range ids {
		raw[i] = id.String()
	}
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE transactions SET status = $1, batch_id = $2, updated_at = NOW()
		WHERE id = ANY($3::uuid[])`, status, batchID, pq.Array(raw))
	return err
}
