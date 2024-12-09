package dto

import (
	"time"

	"github.com/JIeeiroSst/calculate-service/common"
	"github.com/JIeeiroSst/calculate-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
)

type CampaignTypeConfig struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

func (c *CampaignTypeConfig) BuildCreate() model.CampaignTypeConfig {
	id := c.ID
	if len(id) == 0 {
		id = geared_id.GearedStringID()
	}
	return model.CampaignTypeConfig{
		ID:   id,
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
	CampaignContent    []CampaignContent  `json:"campaign_content,omitempty"`
}

func BuildCampaignConfig(c model.CampaignConfig) *CampaignConfig {
	var campaignTypeConfig CampaignTypeConfig
	if c.CampaignTypeConfig != nil {
		campaignTypeConfig = CampaignTypeConfig{
			ID:   c.CampaignTypeConfig.ID,
			Name: c.CampaignTypeConfig.Name,
			Type: c.CampaignTypeConfig.Type,
		}
	}
	campaignContent := make([]CampaignContent, 0)
	for _, v := range c.CampaignContent {
		campaignContent = append(campaignContent, CampaignContent{
			ID:        v.ID,
			Content:   v.Content,
			Value:     v.Value,
			Condition: v.Condition,
		})
	}
	return &CampaignConfig{
		ID:                 c.ID,
		Name:               c.Name,
		Description:        c.Description,
		ClassifyType:       c.ClassifyType,
		CampaignTypeID:     c.CampaignTypeID,
		CampaignTypeConfig: campaignTypeConfig,
		CampaignContent:    campaignContent,
	}
}

type CreateCampaignConfigRequest struct {
	Name            string            `json:"name"`
	Value           float64           `json:"value"`
	Description     string            `json:"description"`
	ClassifyType    string            `json:"classify_type"`
	CampaignTypeID  string            `json:"campaign_type_id"`
	CampaignContent []CampaignContent `json:"campaign_content,omitempty"`
}

type CampaignContent struct {
	ID        string  `json:"id,omitempty"`
	Content   string  `json:"content,omitempty"`
	Value     float64 `json:"value,omitempty"`
	Condition int     `json:"condition,omitempty"`
}

func (c CreateCampaignConfigRequest) Build() model.CampaignConfig {
	campaignContent := make([]model.CampaignContent, 0)
	for _, v := range c.CampaignContent {
		campaignContent = append(campaignContent, model.CampaignContent{
			ID:        v.ID,
			Content:   v.Content,
			Value:     v.Value,
			Condition: v.Condition,
		})
	}
	return model.CampaignConfig{
		ID:              geared_id.GearedStringID(),
		Name:            c.Name,
		Description:     c.Description,
		ClassifyType:    c.ClassifyType,
		CampaignTypeID:  c.CampaignTypeID,
		Status:          common.Draft.Value(),
		UpdatedAt:       time.Now(),
		CreateAdt:       time.Now(),
		CampaignContent: campaignContent,
	}
}

type UpdateCampaignConfigRequest struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Value           float64           `json:"value"`
	Description     string            `json:"description"`
	ClassifyType    string            `json:"classify_type"`
	CampaignTypeID  string            `json:"campaign_type_id"`
	Status          int               `json:"status,omitempty"`
	DeletedAt       time.Time         `json:"deleted_at,omitempty"`
	CampaignContent []CampaignContent `json:"campaign_content,omitempty"`
}

func (c UpdateCampaignConfigRequest) Build() model.CampaignConfig {
	campaignContent := make([]model.CampaignContent, 0)
	for _, v := range c.CampaignContent {
		campaignContent = append(campaignContent, model.CampaignContent{
			ID:        v.ID,
			Content:   v.Content,
			Value:     v.Value,
			Condition: v.Condition,
		})
	}
	return model.CampaignConfig{
		ID:              c.ID,
		Name:            c.Name,
		Description:     c.Description,
		ClassifyType:    c.ClassifyType,
		CampaignTypeID:  c.CampaignTypeID,
		Status:          c.Status,
		DeletedAt:       c.DeletedAt,
		UpdatedAt:       time.Now(),
		CampaignContent: campaignContent,
	}
}

type UserCampaignConfig struct {
	ID             string          `json:"id,omitempty" gorm:"primaryKey"`
	UserID         string          `json:"user_id,omitempty"`
	CampaignID     string          `json:"campaign_id,omitempty"`
	CreateAdt      time.Time       `json:"create_adt,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at,omitempty"`
	TotalAmount    float64         `json:"total_amount,omitempty"`
	DateAt         time.Time       `json:"date_at,omitempty"`
	CampaignConfig *CampaignConfig `json:"campaign_type_config,omitempty"`
}
