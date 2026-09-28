package repository

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"gorm.io/gorm"
)

type dispenseRepository struct {
	db *gorm.DB
}

func NewDispenseRepository(db *gorm.DB) port.DispenseRepository {
	return &dispenseRepository{db: db}
}

func (r *dispenseRepository) Create(ctx context.Context, d *model.Dispense) error {
	return conn(ctx, r.db).Create(d).Error
}

func (r *dispenseRepository) GetByID(ctx context.Context, id int64) (*model.Dispense, error) {
	var d model.Dispense
	if err := conn(ctx, r.db).Preload("Items").First(&d, id).Error; err != nil {
		return nil, translate(err, fmt.Sprintf("dispense %d", id))
	}
	return &d, nil
}

func (r *dispenseRepository) ListByPatient(ctx context.Context, patientRef string, limit, offset int) ([]model.Dispense, error) {
	var dispenses []model.Dispense
	err := conn(ctx, r.db).Preload("Items").Where("patient_ref = ?", patientRef).
		Order("id DESC").Limit(limit).Offset(offset).Find(&dispenses).Error
	return dispenses, err
}
