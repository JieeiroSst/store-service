package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
	"gorm.io/gorm"
)

type adCampaignRepository struct {
	db *gorm.DB
}

func NewAdCampaignRepository(db *gorm.DB) port.AdCampaignRepository {
	return &adCampaignRepository{db: db}
}

func (r *adCampaignRepository) Create(ctx context.Context, campaign *model.AdCampaign) error {
	if err := r.db.WithContext(ctx).Create(campaign).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adCampaignRepository) GetByID(ctx context.Context, id uint) (*model.AdCampaign, error) {
	var campaign model.AdCampaign
	if err := r.db.WithContext(ctx).First(&campaign, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &campaign, nil
}

func (r *adCampaignRepository) Update(ctx context.Context, campaign *model.AdCampaign) error {
	if err := r.db.WithContext(ctx).Model(&model.AdCampaign{}).
		Where("id = ?", campaign.ID).
		Updates(campaign).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adCampaignRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.AdCampaign{}, id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *adCampaignRepository) List(ctx context.Context) ([]model.AdCampaign, error) {
	var campaigns []model.AdCampaign
	if err := r.db.WithContext(ctx).Order("id desc").Find(&campaigns).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return campaigns, nil
}
