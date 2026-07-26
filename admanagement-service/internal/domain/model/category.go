package model

import "time"

type AdCategory struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    *uint     `json:"parent_id,omitempty"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
}

func (AdCategory) TableName() string { return "ad_categories" }
