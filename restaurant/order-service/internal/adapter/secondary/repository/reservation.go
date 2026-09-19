package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/order-service/internal/domain/model"
	"github.com/JIeeiroSst/order-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type reservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) port.ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) Create(ctx context.Context, reservation model.Reservation) error {
	return r.db.WithContext(ctx).Create(&reservation).Error
}

func (r *reservationRepository) Cancel(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&model.Reservation{}).
		Where("id = ?", id).
		Update("status", model.ReservationStatusCancelled).Error
}

func (r *reservationRepository) FindByID(ctx context.Context, id int) (*model.Reservation, error) {
	var reservation model.Reservation
	if err := r.db.WithContext(ctx).Where("id = ?", id).Find(&reservation).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *reservationRepository) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var reservations []*model.Reservation

	r.db.WithContext(ctx).Scopes(logger.Paginate(reservations, &pagination, r.db)).Find(&reservations)
	pagination.Rows = reservations

	return pagination, nil
}

func (r *reservationRepository) FindConflicting(ctx context.Context, tableName string, at time.Time, window time.Duration) ([]model.Reservation, error) {
	var reservations []model.Reservation
	err := r.db.WithContext(ctx).
		Where("table_name = ? AND status = ? AND reserved_at BETWEEN ? AND ?",
			tableName, model.ReservationStatusBooked, at.Add(-window), at.Add(window)).
		Find(&reservations).Error
	if err != nil {
		return nil, err
	}
	return reservations, nil
}
