package application

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type settlementService struct {
	core
}

func NewSettlementService(d Deps) port.SettlementService {
	return &settlementService{core: newCore(d)}
}

func (s *settlementService) CreateBatch(ctx context.Context, merchantID uuid.UUID) (*model.SettlementBatch, error) {
	merchant, err := s.Merchants.GetByID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	var batch *model.SettlementBatch
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		txns, err := s.Transactions.ListUnbatchedCaptures(ctx, merchantID)
		if err != nil {
			return err
		}
		batch, err = model.NewSettlementBatch(merchant, txns, s.now())
		if err != nil {
			return err
		}
		if err := s.Batches.Create(ctx, batch); err != nil {
			return err
		}
		ids := make([]uuid.UUID, 0, len(txns))
		for _, tx := range txns {
			ids = append(ids, tx.ID)
			tx.BatchID = model.NullID(batch.ID)
		}
		return s.Transactions.AssignBatch(ctx, ids, model.TransactionStatusCaptured, batch.ID)
	})
	if err != nil {
		return nil, err
	}
	s.Logger.Info("settlement batch created", zap.String("batch_id", batch.ID.String()), zap.Int("txn_count", batch.TxnCount))
	return batch, nil
}

func (s *settlementService) Clear(ctx context.Context, batchID uuid.UUID) ([]*model.ClearingRecord, error) {
	var records []*model.ClearingRecord
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		batch, err := s.Batches.GetByIDForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		if batch.Status != model.BatchStatusOpen {
			records, err = s.Clearing.ListByBatch(ctx, batchID)
			return err
		}
		txns, err := s.Transactions.ListByBatch(ctx, batchID)
		if err != nil {
			return err
		}
		now := s.now()
		records = model.NetClearing(batchID, txns, now)
		if err := s.Clearing.CreateBulk(ctx, records); err != nil {
			return err
		}
		batch.Status = model.BatchStatusCleared
		return s.Batches.Update(ctx, batch)
	})
	if err != nil {
		return nil, err
	}
	s.Logger.Info("clearing done", zap.String("batch_id", batchID.String()), zap.Int("netted_records", len(records)))
	return records, nil
}

func (s *settlementService) Settle(ctx context.Context, batchID uuid.UUID) error {
	var ids []uuid.UUID
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		batch, err := s.Batches.GetByIDForUpdate(ctx, batchID)
		if err != nil {
			return err
		}
		switch batch.Status {
		case model.BatchStatusSettled:
			return nil
		case model.BatchStatusCleared:
		default:
			return model.ErrBatchState
		}
		txns, err := s.Transactions.ListByBatch(ctx, batchID)
		if err != nil {
			return err
		}
		for _, tx := range txns {
			ids = append(ids, tx.ID)
		}
		if err := s.Transactions.AssignBatch(ctx, ids, model.TransactionStatusSettled, batchID); err != nil {
			return err
		}
		now := s.now()
		batch.Status = model.BatchStatusSettled
		batch.ProcessedAt = &now
		return s.Batches.Update(ctx, batch)
	})
	if err != nil {
		return err
	}
	s.invalidateTransactions(ctx, ids...)
	s.Logger.Info("settlement complete", zap.String("batch_id", batchID.String()), zap.Int("txn_count", len(ids)))
	return nil
}
