package model

import "time"

type AdImpression struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AdID        uint      `json:"ad_id"`
	UserID      *uint     `json:"user_id,omitempty"`
	SessionID   string    `json:"session_id,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	ReferrerURL string    `json:"referrer_url,omitempty"`
	PageURL     string    `json:"page_url,omitempty"`
	PositionID  *uint     `json:"position_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (AdImpression) TableName() string { return "ad_impressions" }
