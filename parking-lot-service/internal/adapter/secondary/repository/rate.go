package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
	"github.com/JIeeiroSst/parking-lot-service/internal/domain/port"
	"gorm.io/gorm"
)

type rateRepository struct {
	db *gorm.DB
}

func NewRateRepository(db *gorm.DB) port.RateRepository {
	return &rateRepository{db: db}
}

func (r *rateRepository) GetActive(ctx context.Context, vehicleType model.VehicleType) (*model.Rate, error) {
	var rate model.Rate
	err := r.db.WithContext(ctx).
		Where("vehicle_type = ? AND effective_from <= ?", vehicleType, time.Now()).
		Order("effective_from DESC").
		First(&rate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &rate, nil
}

func (r *rateRepository) List(ctx context.Context) ([]model.Rate, error) {
	var rates []model.Rate
	err := r.db.WithContext(ctx).Order("vehicle_type, effective_from DESC").Find(&rates).Error
	return rates, err
}
