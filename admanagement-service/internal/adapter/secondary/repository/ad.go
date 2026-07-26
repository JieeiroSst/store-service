package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
)

type adRepository struct {
	db *gorm.DB
}

func NewAdRepository(db *gorm.DB) port.AdRepository {
	return &adRepository{db: db}
}

func (r *adRepository) Create(ctx context.Context, ad *model.Ad) error {
	if err := r.db.WithContext(ctx).Create(ad).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adRepository) GetByID(ctx context.Context, id uint) (*model.Ad, error) {
	var ad model.Ad
	if err := r.db.WithContext(ctx).First(&ad, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &ad, nil
}

func (r *adRepository) Update(ctx context.Context, ad *model.Ad) error {
	if err := r.db.WithContext(ctx).Model(&model.Ad{}).
		Where("id = ?", ad.ID).
		Updates(ad).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Ad{}, id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adRepository) List(ctx context.Context) ([]model.Ad, error) {
	var ads []model.Ad
	if err := r.db.WithContext(ctx).Order("id desc").Find(&ads).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return ads, nil
}

func (r *adRepository) ListByCampaign(ctx context.Context, campaignID uint) ([]model.Ad, error) {
	var ads []model.Ad
	if err := r.db.WithContext(ctx).Where("campaign_id = ?", campaignID).Order("id desc").Find(&ads).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return ads, nil
}
