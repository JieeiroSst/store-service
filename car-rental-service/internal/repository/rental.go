package repository

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RentalRepo struct{ db *gorm.DB }

func (r *RentalRepo) Create(ctx context.Context, rt *model.Rental) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *RentalRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Rental, error) {
	var rt model.Rental
	if err := r.db.WithContext(ctx).First(&rt, "rental_id = ?", id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &rt, nil
}

func (r *RentalRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Rental, error) {
	var rt model.Rental
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&rt, "rental_id = ?", id).Error
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &rt, nil
}

func (r *RentalRepo) Update(ctx context.Context, rt *model.Rental) error {
	return r.db.WithContext(ctx).Save(rt).Error
}

func (r *RentalRepo) ExistsForReservation(ctx context.Context, reservationID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Rental{}).Where("reservation_id = ?", reservationID).Count(&n).Error
	return n > 0, err
}

func (r *RentalRepo) ListByUser(ctx context.Context, userID uuid.UUID, status *model.RentalStatus, offset, limit int) ([]model.Rental, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Rental{}).Where("user_id = ?", userID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []model.Rental
	err := q.Order("pickup_time DESC, rental_id").Offset(offset).Limit(limit).Find(&out).Error
	return out, total, err
}
