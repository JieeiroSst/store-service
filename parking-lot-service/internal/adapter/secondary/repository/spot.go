package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
	"github.com/JIeeiroSst/parking-lot-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type parkingSpotRepository struct {
	db *gorm.DB
}

func NewParkingSpotRepository(db *gorm.DB) port.ParkingSpotRepository {
	return &parkingSpotRepository{db: db}
}

func (r *parkingSpotRepository) FindAndReserveAvailable(ctx context.Context, types []model.SpotType) (*model.ParkingSpot, error) {
	if len(types) == 0 {
		return nil, port.ErrNoSpotAvailable
	}

	var reserved model.ParkingSpot
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, spotType := range types {
			var spot model.ParkingSpot
			err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
				Where("type = ? AND is_available = ?", spotType, true).
				Order("spot_id").
				Limit(1).
				First(&spot).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			if err != nil {
				return err
			}

			if err := tx.Model(&model.ParkingSpot{}).Where("spot_id = ?", spot.SpotID).
				Update("is_available", false).Error; err != nil {
				return err
			}
			spot.IsAvailable = false
			reserved = spot
			return nil
		}
		return port.ErrNoSpotAvailable
	})
	if err != nil {
		return nil, err
	}
	return &reserved, nil
}

func (r *parkingSpotRepository) Release(ctx context.Context, spotID string) error {
	return r.db.WithContext(ctx).Model(&model.ParkingSpot{}).
		Where("spot_id = ?", spotID).
		Update("is_available", true).Error
}

func (r *parkingSpotRepository) CountByType(ctx context.Context) ([]model.SpotAvailability, error) {
	var results []model.SpotAvailability
	err := r.db.WithContext(ctx).Model(&model.ParkingSpot{}).
		Select("type, COUNT(*) FILTER (WHERE is_available) AS available, COUNT(*) AS total").
		Group("type").
		Scan(&results).Error
	return results, err
}
