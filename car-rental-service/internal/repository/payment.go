package repository

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepo struct{ db *gorm.DB }

func (r *PaymentRepo) Create(ctx context.Context, p *model.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PaymentRepo) SumCompleted(ctx context.Context, rentalID uuid.UUID) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&model.Payment{}).
		Where("rental_id = ? AND payment_status = ?", rentalID, model.PaymentStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}
