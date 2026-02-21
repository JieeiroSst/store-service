package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

type transactionRepository struct{ db *sqlx.DB }

func NewTransactionRepository(db *sqlx.DB) ports.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	q := `INSERT INTO transactions
		(id, idempotency_key, wallet_id, merchant_id, acquirer_bank_id, issuer_bank_id,
		 card_id, card_network, amount, currency, fee, type, status, description,
		 authorization_code, batch_id, created_at, updated_at)
		VALUES
		(:id, :idempotency_key, :wallet_id, :merchant_id, :acquirer_bank_id, :issuer_bank_id,
		 :card_id, :card_network, :amount, :currency, :fee, :type, :status, :description,
		 :authorization_code, :batch_id, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, tx)
	return err
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.GetContext(ctx, &tx, `SELECT * FROM transactions WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTransactionNotFound
	}
	return &tx, err
}

func (r *transactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.GetContext(ctx, &tx, `SELECT * FROM transactions WHERE idempotency_key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTransactionNotFound
	}
	return &tx, err
}

func (r *transactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	tx.UpdatedAt = time.Now()
	q := `UPDATE transactions
		SET status=:status, type=:type, authorization_code=:authorization_code,
		    batch_id=:batch_id, updated_at=:updated_at
		WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, tx)
	return err
}

func (r *transactionRepository) ListByWalletID(ctx context.Context, walletID uuid.UUID, page, pageSize int) ([]*domain.Transaction, int64, error) {
	offset := (page - 1) * pageSize
	var txs []*domain.Transaction
	err := r.db.SelectContext(ctx, &txs,
		`SELECT * FROM transactions WHERE wallet_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		walletID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	var count int64
	r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM transactions WHERE wallet_id=$1`, walletID)
	return txs, count, nil
}

func (r *transactionRepository) ListCapturedByMerchant(ctx context.Context, merchantID uuid.UUID, since time.Time) ([]*domain.Transaction, error) {
	var txs []*domain.Transaction
	err := r.db.SelectContext(ctx, &txs,
		`SELECT * FROM transactions
		 WHERE merchant_id=$1 AND status='CAPTURED' AND batch_id='00000000-0000-0000-0000-000000000000'
		   AND created_at >= $2
		 ORDER BY created_at ASC`,
		merchantID, since)
	return txs, err
}

func (r *transactionRepository) UpdateBatch(ctx context.Context, txIDs []uuid.UUID, status domain.TransactionStatus, batchID uuid.UUID) error {
	if len(txIDs) == 0 {
		return nil
	}
	placeholders := make([]string, len(txIDs))
	args := make([]interface{}, 0, len(txIDs)+3)
	args = append(args, status, batchID, time.Now())
	for i, id := range txIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+4)
		args = append(args, id)
	}
	q := fmt.Sprintf(`UPDATE transactions SET status=$1, batch_id=$2, updated_at=$3
	                  WHERE id IN (%s)`, strings.Join(placeholders, ","))
	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}
