package repository

import (
	"context"

	"github.com/JIeeiroSst/calculate-service/model"
	"gorm.io/gorm"
)

type UserCampaignConfigs interface {
	SaveUserCampaignConfig(ctx context.Context, req model.UserCampaignConfig) error
	GetByID(ctx context.Context, id string) (*model.UserCampaignConfig, error)
}

type UserCampaignConfigRepo struct {
	db *gorm.DB
}

func NewUserCampaignConfigRepo(db *gorm.DB) *UserCampaignConfigRepo {
	return &UserCampaignConfigRepo{
		db: db,
	}
}

func (r *UserCampaignConfigRepo) SaveUserCampaignConfig(ctx context.Context, req model.UserCampaignConfig) error {
	user, err := r.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	if user == nil {
		if err := r.db.Create(&req).Error; err != nil {
			return err
		}
		return nil
	}

	if err := r.db.Save(&req).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserCampaignConfigRepo) GetByID(ctx context.Context, id string) (*model.UserCampaignConfig, error) {
	var user model.UserCampaignConfig

	if err := r.db.Where("id = ?", id).Find(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
