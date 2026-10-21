package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const batchColumns = `id, acquirer_id, merchant_id, total_amount, total_fee, txn_count, status, processed_at, created_at`

type settlementBatchRepository struct{ base }

func NewSettlementBatchRepository(db *sqlx.DB) port.SettlementBatchRepository {
	return &settlementBatchRepository{base{db: db}}
}

func (r *settlementBatchRepository) Create(ctx context.Context, b *model.SettlementBatch) error {
	_, err := r.named(ctx, `INSERT INTO settlement_batches (`+batchColumns+`)
		VALUES (:id, :acquirer_id, :merchant_id, :total_amount, :total_fee, :txn_count, :status, :processed_at, :created_at)`, b)
	return err
}

func (r *settlementBatchRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.SettlementBatch, error) {
	var b model.SettlementBatch
	if err := r.get(ctx, &b, model.ErrBatchNotFound,
		`SELECT `+batchColumns+` FROM settlement_batches WHERE id = $1 FOR UPDATE`, id); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *settlementBatchRepository) Update(ctx context.Context, b *model.SettlementBatch) error {
	res, err := r.named(ctx, `UPDATE settlement_batches SET status = :status, processed_at = :processed_at WHERE id = :id`, b)
	if err != nil {
		return err
	}
	return expectOne(res, model.ErrBatchNotFound)
}

type clearingRepository struct{ base }

func NewClearingRepository(db *sqlx.DB) port.ClearingRepository {
	return &clearingRepository{base{db: db}}
}

func (r *clearingRepository) CreateBulk(ctx context.Context, records []*model.ClearingRecord) error {
	for _, rec := range records {
		if _, err := r.named(ctx, `INSERT INTO clearing_records
			(id, batch_id, card_network, acquirer_id, issuer_id, net_amount, cleared_at)
			VALUES (:id, :batch_id, :card_network, :acquirer_id, :issuer_id, :net_amount, :cleared_at)`, rec); err != nil {
			return err
		}
	}
	return nil
}

func (r *clearingRepository) ListByBatch(ctx context.Context, batchID uuid.UUID) ([]*model.ClearingRecord, error) {
	records := make([]*model.ClearingRecord, 0)
	err := sqlx.SelectContext(ctx, r.q(ctx), &records, `SELECT id, batch_id, card_network, acquirer_id, issuer_id, net_amount, cleared_at
		FROM clearing_records WHERE batch_id = $1 ORDER BY cleared_at, id`, batchID)
	return records, err
}
