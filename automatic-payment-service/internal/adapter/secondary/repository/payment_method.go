package repository

import (
	"context"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type paymentMethodRepo struct {
	db *gorm.DB
}

func NewPaymentMethodRepository(db *gorm.DB) port.PaymentMethodRepository {
	return &paymentMethodRepo{db: db}
}

func (r *paymentMethodRepo) Create(ctx context.Context, pm *model.PaymentMethod) error {
	return r.db.WithContext(ctx).Create(pm).Error
}

func (r *paymentMethodRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.PaymentMethod, error) {
	var pm model.PaymentMethod
	err := r.db.WithContext(ctx).Where("payment_method_id = ?", id).First(&pm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &pm, nil
}

func (r *paymentMethodRepo) GetDefaultByUser(ctx context.Context, userID uuid.UUID) (*model.PaymentMethod, error) {
	var pm model.PaymentMethod
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&pm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &pm, nil
}

func (r *paymentMethodRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.PaymentMethod, error) {
	var pms []model.PaymentMethod
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pms).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return pms, nil
}

func (r *paymentMethodRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.PaymentMethod{}, "payment_method_id = ?", id)
	if result.Error != nil {
		return common.ErrDBFailed
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func (r *paymentMethodRepo) ClearDefaultByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.PaymentMethod{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Update("is_default", false).Error
}

func (r *paymentMethodRepo) SetDefault(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&model.PaymentMethod{}).
		Where("payment_method_id = ?", id).
		Update("is_default", true)
	if result.Error != nil {
		return common.ErrDBFailed
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}
