package model

import "gorm.io/datatypes"

type AdPosition struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Width          int            `json:"width"`
	Height         int            `json:"height"`
	MaxFileSize    int            `json:"max_file_size"`
	AllowedFormats datatypes.JSON `json:"allowed_formats,omitempty"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
}

func (AdPosition) TableName() string { return "ad_positions" }
