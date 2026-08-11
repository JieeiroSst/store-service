package model

import "time"

type Recording struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	StreamID        string    `gorm:"column:stream_id;index" json:"streamId"`
	RoomID          string    `gorm:"column:room_id;index" json:"roomId"`
	ObjectKey       string    `gorm:"column:object_key" json:"objectKey"`
	DurationSeconds int       `gorm:"column:duration_seconds" json:"durationSeconds"`
	SizeBytes       int64     `gorm:"column:size_bytes" json:"sizeBytes"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (Recording) TableName() string { return "vod_recordings" }
