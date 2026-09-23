package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) port.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *paymentRepository) GetByID(ctx context.Context, id int64) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).First(&payment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByExternalID(ctx context.Context, externalID string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) List(ctx context.Context, filter port.ListPaymentsFilter) ([]model.Payment, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Payment{})
	if filter.Provider != "" {
		query = query.Where("provider = ?", filter.Provider)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if !filter.From.IsZero() {
		query = query.Where("created_at >= ?", filter.From)
	}
	if !filter.To.IsZero() {
		query = query.Where("created_at <= ?", filter.To)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var payments []model.Payment
	if err := query.Order("created_at desc").Limit(filter.Limit).Offset(filter.Offset).Find(&payments).Error; err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}

func (r *paymentRepository) ApplyRefund(ctx context.Context, id int64, refundedDelta int64, status model.PaymentStatus) error {
	result := r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Updates(map[string]any{
		"refunded_amount": gorm.Expr("refunded_amount + ?", refundedDelta),
		"status":          status,
		"updated_at":      time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id int64, status model.PaymentStatus, externalID, failureReason string) error {
	updates := map[string]any{
		"status":         status,
		"failure_reason": failureReason,
		"updated_at":     time.Now(),
	}
	if externalID != "" {
		updates["external_id"] = externalID
	}

	result := r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}
