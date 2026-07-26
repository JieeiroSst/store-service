package model

import "time"

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

var allowedCampaignTransitions = map[CampaignStatus][]CampaignStatus{
	CampaignStatusDraft:     {CampaignStatusActive, CampaignStatusCancelled},
	CampaignStatusActive:    {CampaignStatusPaused, CampaignStatusCompleted, CampaignStatusCancelled},
	CampaignStatusPaused:    {CampaignStatusActive, CampaignStatusCancelled},
	CampaignStatusCompleted: {},
	CampaignStatusCancelled: {},
}

func (s CampaignStatus) CanTransitionTo(target CampaignStatus) bool {
	for _, next := range allowedCampaignTransitions[s] {
		if next == target {
			return true
		}
	}
	return false
}

type AdCampaign struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Budget      float64        `json:"budget" gorm:"type:decimal(15,2)"`
	StartDate   *time.Time     `json:"start_date,omitempty"`
	EndDate     *time.Time     `json:"end_date,omitempty"`
	Status      CampaignStatus `json:"status" gorm:"type:varchar(20);default:draft"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (AdCampaign) TableName() string { return "ad_campaigns" }
