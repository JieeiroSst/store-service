package dto

import (
	"github.com/JIeeiroSst/calculate-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
)

type CampaignTypeConfig struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

func (c *CampaignTypeConfig) BuildCreate() model.CampaignTypeConfig {
	return model.CampaignTypeConfig{
		ID:   geared_id.GearedStringID(),
		Type: c.Type,
		Name: c.Name,
	}
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

type CreateCampaignConfigRequest struct {
}
