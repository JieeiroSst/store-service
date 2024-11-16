package repository

import "gorm.io/gorm"

type Repository struct {
	UserCampaignConfigs
	CampaignConfigs
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		UserCampaignConfigs: NewUserCampaignConfigRepo(db),
		CampaignConfigs:     NewCampaignConfigRepo(db),
	}
}
