package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/address-country-service/common"
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
	"gorm.io/gorm"
)

type shippingOrderanditemchangesCountryRepository struct {
	db *gorm.DB
}

func NewShippingOrderanditemchangesCountryRepository(db *gorm.DB) port.ShippingOrderanditemchangesCountryRepository {
	return &shippingOrderanditemchangesCountryRepository{db: db}
}

func (r *shippingOrderanditemchangesCountryRepository) Create(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *shippingOrderanditemchangesCountryRepository) GetByID(ctx context.Context, id int) (*model.ShippingOrderanditemchangesCountry, error) {
	var entry model.ShippingOrderanditemchangesCountry
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &entry, nil
}

func (r *shippingOrderanditemchangesCountryRepository) Update(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error {
	if err := r.db.WithContext(ctx).Model(&model.ShippingOrderanditemchangesCountry{}).
		Where("id = ?", entry.ID).
		Updates(entry).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *shippingOrderanditemchangesCountryRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ShippingOrderanditemchangesCountry{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *shippingOrderanditemchangesCountryRepository) List(ctx context.Context) ([]model.ShippingOrderanditemchangesCountry, error) {
	var entries []model.ShippingOrderanditemchangesCountry
	if err := r.db.WithContext(ctx).Find(&entries).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return entries, nil
}
