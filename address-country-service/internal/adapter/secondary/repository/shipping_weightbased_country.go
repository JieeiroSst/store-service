package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/address-country-service/common"
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
	"gorm.io/gorm"
)

type shippingWeightbasedCountryRepository struct {
	db *gorm.DB
}

func NewShippingWeightbasedCountryRepository(db *gorm.DB) port.ShippingWeightbasedCountryRepository {
	return &shippingWeightbasedCountryRepository{db: db}
}

func (r *shippingWeightbasedCountryRepository) Create(ctx context.Context, entry *model.ShippingWeightbasedCountry) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *shippingWeightbasedCountryRepository) GetByID(ctx context.Context, id int) (*model.ShippingWeightbasedCountry, error) {
	var entry model.ShippingWeightbasedCountry
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &entry, nil
}

func (r *shippingWeightbasedCountryRepository) Update(ctx context.Context, entry *model.ShippingWeightbasedCountry) error {
	if err := r.db.WithContext(ctx).Model(&model.ShippingWeightbasedCountry{}).
		Where("id = ?", entry.ID).
		Updates(entry).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *shippingWeightbasedCountryRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ShippingWeightbasedCountry{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *shippingWeightbasedCountryRepository) List(ctx context.Context) ([]model.ShippingWeightbasedCountry, error) {
	var entries []model.ShippingWeightbasedCountry
	if err := r.db.WithContext(ctx).Find(&entries).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return entries, nil
}
