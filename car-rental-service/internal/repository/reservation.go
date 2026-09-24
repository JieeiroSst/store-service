package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReservationRepo struct{ db *gorm.DB }

func (r *ReservationRepo) Create(ctx context.Context, res *model.Reservation) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *ReservationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	var res model.Reservation
	if err := r.db.WithContext(ctx).First(&res, "reservation_id = ?", id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &res, nil
}

func (r *ReservationRepo) Update(ctx context.Context, res *model.Reservation) error {
	return r.db.WithContext(ctx).Save(res).Error
}

func (r *ReservationRepo) HasOverlap(ctx context.Context, vehicleID uuid.UUID, start, end time.Time, excludeID uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Reservation{}).
		Where("vehicle_id = ? AND status IN ? AND start_time < ? AND end_time > ?",
			vehicleID, []model.ReservationStatus{model.ReservationStatusPending, model.ReservationStatusConfirmed}, end, start)
	if excludeID != uuid.Nil {
		q = q.Where("reservation_id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}

func (r *ReservationRepo) ListByUser(ctx context.Context, userID uuid.UUID, status *model.ReservationStatus, offset, limit int) ([]model.Reservation, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Reservation{}).Where("user_id = ?", userID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []model.Reservation
	err := q.Order("start_time DESC, reservation_id").Offset(offset).Limit(limit).Find(&out).Error
	return out, total, err
}
