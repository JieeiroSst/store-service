package repository

import "gorm.io/gorm"

type CampaignConfigs interface {
}

type CampaignConfigRepo struct {
	db *gorm.DB
}

func NewCampaignConfigRepo(db *gorm.DB) *CampaignConfigRepo {
	return &CampaignConfigRepo{
		db: db,
	}
}

