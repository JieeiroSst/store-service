package repository

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"gorm.io/gorm"
)

type warehouseRepository struct {
	db *gorm.DB
}

func NewWarehouseRepository(db *gorm.DB) port.WarehouseRepository {
	return &warehouseRepository{db: db}
}

func (r *warehouseRepository) Create(ctx context.Context, w *model.Warehouse) error {
	return translate(conn(ctx, r.db).Create(w).Error, "warehouse "+w.Code)
}

func (r *warehouseRepository) Save(ctx context.Context, w *model.Warehouse) error {
	return conn(ctx, r.db).Select("Active", "UpdatedAt").Updates(w).Error
}

func (r *warehouseRepository) GetByID(ctx context.Context, id int64) (*model.Warehouse, error) {
	var w model.Warehouse
	if err := conn(ctx, r.db).First(&w, id).Error; err != nil {
		return nil, translate(err, fmt.Sprintf("warehouse %d", id))
	}
	return &w, nil
}

func (r *warehouseRepository) List(ctx context.Context) ([]model.Warehouse, error) {
	var ws []model.Warehouse
	err := conn(ctx, r.db).Order("code ASC").Find(&ws).Error
	return ws, err
}

func (r *warehouseRepository) ListActive(ctx context.Context, ids []int64) ([]model.Warehouse, error) {
	q := conn(ctx, r.db).Where("active = ?", true)
	if len(ids) > 0 {
		q = q.Where("id IN ?", ids)
	}
	var ws []model.Warehouse
	err := q.Order("id ASC").Find(&ws).Error
	return ws, err
}
