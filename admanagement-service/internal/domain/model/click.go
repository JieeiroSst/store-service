package model

import "time"

type AdClick struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AdID         uint      `json:"ad_id"`
	ImpressionID *uint     `json:"impression_id,omitempty"`
	UserID       *uint     `json:"user_id,omitempty"`
	SessionID    string    `json:"session_id,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	ReferrerURL  string    `json:"referrer_url,omitempty"`
	TargetURL    string    `json:"target_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (AdClick) TableName() string { return "ad_clicks" }
