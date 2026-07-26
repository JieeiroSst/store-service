package model

import "time"

type AdType string

const (
	AdTypeImage  AdType = "image"
	AdTypeVideo  AdType = "video"
	AdTypeBanner AdType = "banner"
	AdTypeText   AdType = "text"
	AdTypeLink   AdType = "link"
)

type Ad struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	CampaignID  uint       `json:"campaign_id"`
	CategoryID  *uint      `json:"category_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AdType      AdType     `json:"ad_type"`
	ContentURL  string     `json:"content_url,omitempty"`
	TargetURL   string     `json:"target_url,omitempty"`
	FilePath    string     `json:"file_path,omitempty"`
	FileSize    int        `json:"file_size,omitempty"`
	MimeType    string     `json:"mime_type,omitempty"`
	Duration    int        `json:"duration,omitempty"`
	Priority    int        `json:"priority"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Ad) TableName() string { return "ads" }

func (a Ad) IsWithinSchedule(t time.Time) bool {
	if a.StartDate != nil && t.Before(*a.StartDate) {
		return false
	}
	if a.EndDate != nil && t.After(*a.EndDate) {
		return false
	}
	return true
}
