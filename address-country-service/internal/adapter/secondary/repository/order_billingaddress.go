package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/address-country-service/common"
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
	"gorm.io/gorm"
)

type orderBillingaddressRepository struct {
	db *gorm.DB
}

func NewOrderBillingaddressRepository(db *gorm.DB) port.OrderBillingaddressRepository {
	return &orderBillingaddressRepository{db: db}
}

func (r *orderBillingaddressRepository) Create(ctx context.Context, address *model.OrderBillingaddress) error {
	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *orderBillingaddressRepository) GetByID(ctx context.Context, id int) (*model.OrderBillingaddress, error) {
	var address model.OrderBillingaddress
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &address, nil
}

func (r *orderBillingaddressRepository) Update(ctx context.Context, address *model.OrderBillingaddress) error {
	if err := r.db.WithContext(ctx).Model(&model.OrderBillingaddress{}).
		Where("id = ?", address.ID).
		Updates(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *orderBillingaddressRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.OrderBillingaddress{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *orderBillingaddressRepository) List(ctx context.Context) ([]model.OrderBillingaddress, error) {
	var addresses []model.OrderBillingaddress
	if err := r.db.WithContext(ctx).Find(&addresses).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return addresses, nil
}
