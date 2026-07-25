package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/address-country-service/common"
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
	"gorm.io/gorm"
)

type addressCountryRepository struct {
	db *gorm.DB
}

func NewAddressCountryRepository(db *gorm.DB) port.AddressCountryRepository {
	return &addressCountryRepository{db: db}
}

func (r *addressCountryRepository) Create(ctx context.Context, address *model.AddressCountry) error {
	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *addressCountryRepository) GetByID(ctx context.Context, id string) (*model.AddressCountry, error) {
	var address model.AddressCountry
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &address, nil
}

func (r *addressCountryRepository) Update(ctx context.Context, address *model.AddressCountry) error {
	if err := r.db.WithContext(ctx).Model(&model.AddressCountry{}).
		Where("id = ?", address.Id).
		Updates(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *addressCountryRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AddressCountry{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *addressCountryRepository) List(ctx context.Context) ([]model.AddressCountry, error) {
	var addresses []model.AddressCountry
	if err := r.db.WithContext(ctx).Find(&addresses).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return addresses, nil
}
