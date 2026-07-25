package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/address-country-service/common"
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/internal/domain/port"
	"gorm.io/gorm"
)

type partneraddressRepository struct {
	db *gorm.DB
}

func NewPartneraddressRepository(db *gorm.DB) port.PartneraddressRepository {
	return &partneraddressRepository{db: db}
}

func (r *partneraddressRepository) Create(ctx context.Context, address *model.Partneraddress) error {
	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *partneraddressRepository) GetByID(ctx context.Context, id int) (*model.Partneraddress, error) {
	var address model.Partneraddress
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &address, nil
}

func (r *partneraddressRepository) Update(ctx context.Context, address *model.Partneraddress) error {
	if err := r.db.WithContext(ctx).Model(&model.Partneraddress{}).
		Where("id = ?", address.ID).
		Updates(address).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *partneraddressRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Partneraddress{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *partneraddressRepository) List(ctx context.Context) ([]model.Partneraddress, error) {
	var addresses []model.Partneraddress
	if err := r.db.WithContext(ctx).Find(&addresses).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return addresses, nil
}
