package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/address-country-service/common"
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
	"gorm.io/gorm"
)

type orderShippingaddressRepository struct {
	db *gorm.DB
}

func NewOrderShippingaddressRepository(db *gorm.DB) port.OrderShippingaddressRepository {
	return &orderShippingaddressRepository{db: db}
}

func (r *orderShippingaddressRepository) Create(ctx context.Context, address *model.OrderShippingaddress) error {
	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *orderShippingaddressRepository) GetByID(ctx context.Context, id int) (*model.OrderShippingaddress, error) {
	var address model.OrderShippingaddress
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &address, nil
}

func (r *orderShippingaddressRepository) Update(ctx context.Context, address *model.OrderShippingaddress) error {
	if err := r.db.WithContext(ctx).Model(&model.OrderShippingaddress{}).
		Where("id = ?", address.ID).
		Updates(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *orderShippingaddressRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.OrderShippingaddress{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *orderShippingaddressRepository) List(ctx context.Context) ([]model.OrderShippingaddress, error) {
	var addresses []model.OrderShippingaddress
	if err := r.db.WithContext(ctx).Find(&addresses).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return addresses, nil
}
