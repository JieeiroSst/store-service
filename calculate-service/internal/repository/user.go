package repository

import "gorm.io/gorm"

type UserCampaignConfigs interface {
}

type UserCampaignConfigRepo struct {
	db *gorm.DB
}

func NewUserCampaignConfigRepo(db *gorm.DB) *UserCampaignConfigRepo {
	return &UserCampaignConfigRepo{
		db: db,
	}
}

