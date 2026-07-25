package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/address-country-service/common"
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
	"gorm.io/gorm"
)

type userAddressRepository struct {
	db *gorm.DB
}

func NewUserAddressRepository(db *gorm.DB) port.UserAddressRepository {
	return &userAddressRepository{db: db}
}

func (r *userAddressRepository) Create(ctx context.Context, address *model.UserAddress) error {
	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *userAddressRepository) GetByID(ctx context.Context, id int) (*model.UserAddress, error) {
	var address model.UserAddress
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &address, nil
}

func (r *userAddressRepository) Update(ctx context.Context, address *model.UserAddress) error {
	if err := r.db.WithContext(ctx).Model(&model.UserAddress{}).
		Where("id = ?", address.ID).
		Updates(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *userAddressRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.UserAddress{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *userAddressRepository) List(ctx context.Context) ([]model.UserAddress, error) {
	var addresses []model.UserAddress
	if err := r.db.WithContext(ctx).Find(&addresses).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return addresses, nil
}
