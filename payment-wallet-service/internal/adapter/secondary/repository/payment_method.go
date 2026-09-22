package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"gorm.io/gorm"
)

type paymentMethodRepository struct {
	db *gorm.DB
}

func NewPaymentMethodRepository(db *gorm.DB) port.PaymentMethodRepository {
	return &paymentMethodRepository{db: db}
}

func (r *paymentMethodRepository) Create(ctx context.Context, method *model.PaymentMethod) error {
	method.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(method).Error
}

func (r *paymentMethodRepository) GetByID(ctx context.Context, id string) (*model.PaymentMethod, error) {
	var method model.PaymentMethod
	if err := r.db.WithContext(ctx).Where("payment_method_id = ?", id).First(&method).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &method, nil
}

func (r *paymentMethodRepository) ListByUser(ctx context.Context, userID string) ([]model.PaymentMethod, error) {
	var methods []model.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("is_default DESC, created_at DESC").
		Find(&methods).Error
	return methods, err
}

func (r *paymentMethodRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.PaymentMethod{}).
		Where("payment_method_id = ?", id).
		Update("is_active", false).Error
}

func (r *paymentMethodRepository) SetDefault(ctx context.Context, id string) (*model.PaymentMethod, error) {
	var updated model.PaymentMethod
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var method model.PaymentMethod
		if err := tx.Where("payment_method_id = ?", id).First(&method).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return port.ErrNotFound
			}
			return err
		}

		if err := tx.Model(&model.PaymentMethod{}).
			Where("user_id = ? AND payment_method_id <> ?", method.UserID, id).
			Update("is_default", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.PaymentMethod{}).
			Where("payment_method_id = ?", id).
			Update("is_default", true).Error; err != nil {
			return err
		}

		method.IsDefault = true
		updated = method
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
