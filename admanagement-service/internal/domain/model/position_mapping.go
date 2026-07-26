package model

import "time"

type AdPositionMapping struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	AdID       uint      `json:"ad_id"`
	PositionID uint      `json:"position_id"`
	Weight     int       `json:"weight" gorm:"default:1"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AdPositionMapping) TableName() string { return "ad_position_mappings" }
