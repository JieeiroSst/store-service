package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

const maxExpiringDays = 365

type inventoryService struct {
	medicines port.MedicineRepository
	batches   port.BatchRepository
	movements port.MovementRepository
	tx        port.TxManager
	clock     port.Clock
}

func NewInventoryService(
	medicines port.MedicineRepository,
	batches port.BatchRepository,
	movements port.MovementRepository,
	tx port.TxManager,
	clock port.Clock,
) port.InventoryUsecase {
	return &inventoryService{medicines: medicines, batches: batches, movements: movements, tx: tx, clock: clock}
}

func (s *inventoryService) ReceiveBatch(ctx context.Context, in port.ReceiveBatchInput) (*model.Batch, error) {
	in.BatchNumber = strings.TrimSpace(in.BatchNumber)
	switch {
	case in.BatchNumber == "":
		return nil, invalid("batch_number is required")
	case in.Quantity <= 0:
		return nil, invalid("quantity must be greater than zero")
	case in.UnitCost < 0:
		return nil, invalid("unit_cost must not be negative")
	case in.ReceivedBy == "":
		return nil, invalid("received_by is required")
	case in.ExpiryDate <= s.clock.Today():
		return nil, invalid("cannot receive a batch that is already expired (expiry %s)", in.ExpiryDate)
	}
	if _, err := s.medicines.GetByID(ctx, in.MedicineID); err != nil {
		return nil, err
	}

	var batch *model.Batch
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		existing, err := s.batches.FindByNumberForUpdate(ctx, in.MedicineID, in.BatchNumber)
		switch {
		case err == nil:
			if existing.ExpiryDate != in.ExpiryDate {
				return fmt.Errorf("%w: batch %s already exists with expiry %s", port.ErrConflict, in.BatchNumber, existing.ExpiryDate)
			}
			existing.Quantity += in.Quantity
			existing.ReceivedQuantity += in.Quantity
			if err := s.batches.Update(ctx, existing); err != nil {
				return err
			}
			batch = existing
		case errors.Is(err, port.ErrNotFound):
			batch = &model.Batch{
				MedicineID:       in.MedicineID,
				BatchNumber:      in.BatchNumber,
				ExpiryDate:       in.ExpiryDate,
				Quantity:         in.Quantity,
				ReceivedQuantity: in.Quantity,
				UnitCost:         in.UnitCost,
				Supplier:         strings.TrimSpace(in.Supplier),
				ReceivedAt:       s.clock.Now(),
			}
			if err := s.batches.Create(ctx, batch); err != nil {
				return err
			}
		default:
			return err
		}

		return s.movements.Create(ctx, []model.StockMovement{{
			MedicineID: in.MedicineID,
			BatchID:    batch.ID,
			Type:       model.MovementReceive,
			Quantity:   in.Quantity,
			Reference:  in.BatchNumber,
			CreatedBy:  in.ReceivedBy,
		}})
	})
	if err != nil {
		return nil, err
	}
	return batch, nil
}

func (s *inventoryService) AdjustBatch(ctx context.Context, in port.AdjustBatchInput) (*model.Batch, error) {
	in.Reason = strings.TrimSpace(in.Reason)
	switch {
	case in.NewQuantity < 0:
		return nil, invalid("quantity must not be negative")
	case in.Reason == "":
		return nil, invalid("reason is required")
	case in.AdjustedBy == "":
		return nil, invalid("adjusted_by is required")
	}

	var batch *model.Batch
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		b, err := s.batches.GetByIDForUpdate(ctx, in.BatchID)
		if err != nil {
			return err
		}
		delta := in.NewQuantity - b.Quantity
		if delta == 0 {
			batch = b
			return nil
		}
		b.Quantity = in.NewQuantity
		if err := s.batches.Update(ctx, b); err != nil {
			return err
		}
		batch = b
		return s.movements.Create(ctx, []model.StockMovement{{
			MedicineID: b.MedicineID,
			BatchID:    b.ID,
			Type:       model.MovementAdjust,
			Quantity:   delta,
			Reason:     in.Reason,
			CreatedBy:  in.AdjustedBy,
		}})
	})
	if err != nil {
		return nil, err
	}
	return batch, nil
}

func (s *inventoryService) GetStock(ctx context.Context, medicineID int64) (*model.StockSummary, error) {
	if _, err := s.medicines.GetByID(ctx, medicineID); err != nil {
		return nil, err
	}
	batches, err := s.batches.ListByMedicine(ctx, medicineID)
	if err != nil {
		return nil, err
	}
	summary := model.SummarizeStock(medicineID, batches, s.clock.Today())
	return &summary, nil
}

func (s *inventoryService) ListExpiring(ctx context.Context, withinDays int) ([]model.Batch, error) {
	if withinDays <= 0 || withinDays > maxExpiringDays {
		return nil, invalid("days must be between 1 and %d", maxExpiringDays)
	}
	today := s.clock.Today()
	return s.batches.ListExpiring(ctx, today, today.AddDays(withinDays))
}

func (s *inventoryService) ListLowStock(ctx context.Context) ([]model.LowStockItem, error) {
	return s.medicines.ListLowStock(ctx, s.clock.Today())
}

func (s *inventoryService) DisposeExpired(ctx context.Context, disposedBy string) (*port.DisposeResult, error) {
	if strings.TrimSpace(disposedBy) == "" {
		return nil, invalid("disposed_by is required")
	}

	result := &port.DisposeResult{}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		expired, err := s.batches.ListExpiredForUpdate(ctx, s.clock.Today())
		if err != nil {
			return err
		}
		movements := make([]model.StockMovement, 0, len(expired))
		for i := range expired {
			b := &expired[i]
			movements = append(movements, model.StockMovement{
				MedicineID: b.MedicineID,
				BatchID:    b.ID,
				Type:       model.MovementDispose,
				Quantity:   -b.Quantity,
				Reason:     fmt.Sprintf("expired on %s", b.ExpiryDate),
				CreatedBy:  disposedBy,
			})
			result.Batches++
			result.Units += b.Quantity
			b.Quantity = 0
			if err := s.batches.Update(ctx, b); err != nil {
				return err
			}
		}
		if len(movements) == 0 {
			return nil
		}
		return s.movements.Create(ctx, movements)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *inventoryService) ListMovements(ctx context.Context, medicineID int64, in port.ListInput) ([]model.StockMovement, error) {
	limit, offset := normalizeList(in)
	return s.movements.ListByMedicine(ctx, medicineID, limit, offset)
}
