package repository

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
	"github.com/JIeeiroSst/accounting-service/internal/domain/port"
	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) port.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment model.Payment) error {
	return r.db.WithContext(ctx).Create(&payment).Error
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, orderID int, status string) error {
	return r.db.WithContext(ctx).Model(&model.Payment{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID int) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}
