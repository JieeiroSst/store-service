package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)


type settlementBatchRepository struct{ db *sqlx.DB }

func NewSettlementBatchRepository(db *sqlx.DB) ports.SettlementBatchRepository {
	return &settlementBatchRepository{db: db}
}

func (r *settlementBatchRepository) Create(ctx context.Context, batch *domain.SettlementBatch) error {
	q := `INSERT INTO settlement_batches
		(id, acquirer_id, merchant_id, total_amount, total_fee, txn_count, status, created_at)
		VALUES (:id, :acquirer_id, :merchant_id, :total_amount, :total_fee, :txn_count, :status, :created_at)`
	_, err := r.db.NamedExecContext(ctx, q, batch)
	return err
}

func (r *settlementBatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SettlementBatch, error) {
	var b domain.SettlementBatch
	err := r.db.GetContext(ctx, &b, `SELECT * FROM settlement_batches WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrBatchNotFound
	}
	if err != nil {
		return nil, err
	}
	var txs []domain.Transaction
	r.db.SelectContext(ctx, &txs, `SELECT * FROM transactions WHERE batch_id = $1`, id)
	b.Transactions = txs
	return &b, nil
}

func (r *settlementBatchRepository) Update(ctx context.Context, batch *domain.SettlementBatch) error {
	q := `UPDATE settlement_batches SET status=:status, processed_at=:processed_at WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, batch)
	return err
}

type clearingRepository struct{ db *sqlx.DB }

func NewClearingRepository(db *sqlx.DB) ports.ClearingRepository {
	return &clearingRepository{db: db}
}

func (r *clearingRepository) CreateBulk(ctx context.Context, records []*domain.ClearingRecord) error {
	if len(records) == 0 {
		return nil
	}
	q := `INSERT INTO clearing_records (id, batch_id, card_network, acquirer_id, issuer_id, net_amount, cleared_at)
	      VALUES (:id, :batch_id, :card_network, :acquirer_id, :issuer_id, :net_amount, :cleared_at)`
	for _, rec := range records {
		if _, err := r.db.NamedExecContext(ctx, q, rec); err != nil {
			return err
		}
	}
	return nil
}

func (r *clearingRepository) GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.ClearingRecord, error) {
	var records []*domain.ClearingRecord
	err := r.db.SelectContext(ctx, &records, `SELECT * FROM clearing_records WHERE batch_id = $1`, batchID)
	return records, err
}
