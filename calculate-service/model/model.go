package model

type CampaignTypeConfig struct {
	ID   string
	Type string
	Name string
}

type CampaignConfig struct {
	ID             string
	Name           string
	Value          float64
	Description    string
	ClassifyType   string
	CampaignTypeID string
}

