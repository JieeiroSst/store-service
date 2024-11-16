package model

import "time"

type CampaignTypeConfig struct {
	ID   string `gorm:"primaryKey" json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type CampaignConfig struct {
	ID                 string              `json:"id,omitempty" gorm:"primaryKey"`
	Name               string              `json:"name,omitempty" gorm:"index"`
	Value              float64             `json:"value,omitempty" gorm:"index"`
	Description        string              `json:"description,omitempty"`
	ClassifyType       string              `json:"classify_type,omitempty"`
	CampaignTypeID     string              `json:"campaign_type_id,omitempty"`
	Status             int                 `json:"status,omitempty"`
	CreateAdt          time.Time           `json:"create_adt,omitempty"`
	UpdatedAt          time.Time           `json:"updated_at,omitempty"`
	DeletedAt          time.Time           `json:"deleted_at,omitempty"`
	CampaignTypeConfig *CampaignTypeConfig `json:"campaign_type_config,omitempty" gorm:"foreignKey:CampaignTypeID"`
}

type UserCampaignConfig struct {
	ID             string          `json:"id,omitempty" gorm:"primaryKey"`
	UserID         string          `json:"user_id,omitempty"`
	CampaignID     string          `json:"campaign_id,omitempty"`
	CreateAdt      time.Time       `json:"create_adt,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at,omitempty"`
	TotalAmount    float64         `json:"total_amount,omitempty"`
	DateAt         time.Time       `json:"date_at,omitempty"`
	CampaignConfig *CampaignConfig `json:"campaign_type_config,omitempty" gorm:"foreignKey:CampaignID"`
}
