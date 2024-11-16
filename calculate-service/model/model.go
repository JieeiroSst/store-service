package model

type CampaignTypeConfig struct {
	ID   string `gorm:"primaryKey" json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type CampaignConfig struct {
	ID                 string             `json:"id" gorm:"primaryKey"`
	Name               string             `json:"name" gorm:"index"`
	Value              float64            `json:"value" gorm:"index"`
	Description        string             `json:"description"`
	ClassifyType       string             `json:"classify_type"`
	CampaignTypeID     string             `json:"campaign_type_id"`
	CampaignTypeConfig CampaignTypeConfig `json:"campaign_type_config" gorm:"foreignKey:CampaignTypeID"`
}
