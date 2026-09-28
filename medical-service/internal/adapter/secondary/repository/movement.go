package repository

import (
	"context"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"gorm.io/gorm"
)

type movementRepository struct {
	db *gorm.DB
}

func NewMovementRepository(db *gorm.DB) port.MovementRepository {
	return &movementRepository{db: db}
}

func (r *movementRepository) Create(ctx context.Context, movements []model.StockMovement) error {
	if len(movements) == 0 {
		return nil
	}
	return conn(ctx, r.db).Create(&movements).Error
}

func (r *movementRepository) ListByMedicine(ctx context.Context, medicineID int64, limit, offset int) ([]model.StockMovement, error) {
	var movements []model.StockMovement
	err := conn(ctx, r.db).Where("medicine_id = ?", medicineID).
		Order("id DESC").Limit(limit).Offset(offset).Find(&movements).Error
	return movements, err
}
