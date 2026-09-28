package repository

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewRepo struct{ db *gorm.DB }

func (r *ReviewRepo) Create(ctx context.Context, rv *model.Review) error {
	return r.db.WithContext(ctx).Create(rv).Error
}

func (r *ReviewRepo) ExistsForRental(ctx context.Context, rentalID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Review{}).Where("rental_id = ?", rentalID).Count(&n).Error
	return n > 0, err
}

func (r *ReviewRepo) ListByVehicle(ctx context.Context, vehicleID uuid.UUID, offset, limit int) ([]model.Review, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Review{}).Where("vehicle_id = ?", vehicleID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []model.Review
	if err := q.Order("created_at DESC, review_id").Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
