package model

import "time"

type AdPerformanceSummary struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AdID        uint      `json:"ad_id"`
	Date        time.Time `json:"date" gorm:"type:date"`
	Impressions int       `json:"impressions"`
	Clicks      int       `json:"clicks"`
	CTR         float64   `json:"ctr" gorm:"type:decimal(5,4)"`
	Cost        float64   `json:"cost" gorm:"type:decimal(15,2)"`
	Revenue     float64   `json:"revenue" gorm:"type:decimal(15,2)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (AdPerformanceSummary) TableName() string { return "ad_performance_summary" }

type CampaignPerformance struct {
	CampaignID  uint    `json:"campaign_id"`
	Impressions int     `json:"impressions"`
	Clicks      int     `json:"clicks"`
	CTR         float64 `json:"ctr"`
	Cost        float64 `json:"cost"`
	Revenue     float64 `json:"revenue"`
}
