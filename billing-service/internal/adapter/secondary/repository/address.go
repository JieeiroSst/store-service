package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/billing-service/common"
	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
	"gorm.io/gorm"
)

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) port.AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, address *model.Address) error {
	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *addressRepository) GetByID(ctx context.Context, id int) (*model.Address, error) {
	var address model.Address
	if err := r.db.WithContext(ctx).Where("address_id = ?", id).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &address, nil
}

func (r *addressRepository) Update(ctx context.Context, address *model.Address) error {
	if err := r.db.WithContext(ctx).Model(&model.Address{}).
		Where("address_id = ?", address.AddressID).
		Updates(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *addressRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("address_id = ?", id).Delete(&model.Address{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *addressRepository) List(ctx context.Context) ([]model.Address, error) {
	var addresses []model.Address
	if err := r.db.WithContext(ctx).Find(&addresses).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return addresses, nil
}
