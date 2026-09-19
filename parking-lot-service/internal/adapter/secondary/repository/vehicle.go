package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
	"github.com/JIeeiroSst/parking-lot-service/internal/domain/port"
	"gorm.io/gorm"
)

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) port.VehicleRepository {
	return &vehicleRepository{db: db}
}

func (r *vehicleRepository) GetOrCreate(ctx context.Context, plate string, vehicleType model.VehicleType) (*model.Vehicle, error) {
	var vehicle model.Vehicle
	err := r.db.WithContext(ctx).Where("license_plate = ?", plate).First(&vehicle).Error
	if err == nil {
		return &vehicle, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	vehicle = model.Vehicle{LicensePlate: plate, Type: vehicleType}
	if err := r.db.WithContext(ctx).Clauses(onConflictDoNothing("license_plate")).Create(&vehicle).Error; err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) Get(ctx context.Context, plate string) (*model.Vehicle, error) {
	var vehicle model.Vehicle
	if err := r.db.WithContext(ctx).Where("license_plate = ?", plate).First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &vehicle, nil
}
