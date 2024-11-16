package dto

type CampaignTypeConfig struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type CampaignConfig struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Value              float64            `json:"value"`
	Description        string             `json:"description"`
	ClassifyType       string             `json:"classify_type"`
	CampaignTypeID     string             `json:"campaign_type_id"`
	CampaignTypeConfig CampaignTypeConfig `json:"campaign_type_config"`
}
