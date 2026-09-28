package repository

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"gorm.io/gorm"
)

type batchRepository struct {
	db *gorm.DB
}

func NewBatchRepository(db *gorm.DB) port.BatchRepository {
	return &batchRepository{db: db}
}

func (r *batchRepository) Create(ctx context.Context, b *model.Batch) error {
	return translate(conn(ctx, r.db).Create(b).Error, "batch "+b.BatchNumber)
}

func (r *batchRepository) Update(ctx context.Context, b *model.Batch) error {
	return conn(ctx, r.db).Model(b).Select("Quantity", "ReceivedQuantity", "UpdatedAt").Updates(b).Error
}

func (r *batchRepository) GetByIDForUpdate(ctx context.Context, id int64) (*model.Batch, error) {
	var b model.Batch
	if err := conn(ctx, r.db).Clauses(forUpdate).First(&b, id).Error; err != nil {
		return nil, translate(err, fmt.Sprintf("batch %d", id))
	}
	return &b, nil
}

func (r *batchRepository) FindByNumberForUpdate(ctx context.Context, medicineID int64, batchNumber string) (*model.Batch, error) {
	var b model.Batch
	err := conn(ctx, r.db).Clauses(forUpdate).
		Where("medicine_id = ? AND batch_number = ?", medicineID, batchNumber).
		First(&b).Error
	if err != nil {
		return nil, translate(err, "batch "+batchNumber)
	}
	return &b, nil
}

func (r *batchRepository) ListByMedicine(ctx context.Context, medicineID int64) ([]model.Batch, error) {
	var batches []model.Batch
	err := conn(ctx, r.db).Where("medicine_id = ?", medicineID).
		Order("expiry_date ASC, id ASC").Find(&batches).Error
	return batches, err
}

func (r *batchRepository) ListUsableForUpdate(ctx context.Context, medicineID int64, today model.Date) ([]model.Batch, error) {
	var batches []model.Batch
	err := conn(ctx, r.db).Clauses(forUpdate).
		Where("medicine_id = ? AND quantity > 0 AND expiry_date > ?", medicineID, today).
		Order("expiry_date ASC, id ASC").Find(&batches).Error
	return batches, err
}

func (r *batchRepository) ListExpiredForUpdate(ctx context.Context, today model.Date) ([]model.Batch, error) {
	var batches []model.Batch
	err := conn(ctx, r.db).Clauses(forUpdate).
		Where("quantity > 0 AND expiry_date <= ?", today).
		Order("id ASC").Find(&batches).Error
	return batches, err
}

func (r *batchRepository) ListExpiring(ctx context.Context, today, until model.Date) ([]model.Batch, error) {
	var batches []model.Batch
	err := conn(ctx, r.db).
		Where("quantity > 0 AND expiry_date > ? AND expiry_date <= ?", today, until).
		Order("expiry_date ASC, id ASC").Find(&batches).Error
	return batches, err
}
