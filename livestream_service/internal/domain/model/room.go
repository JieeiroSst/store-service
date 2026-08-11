package model

import "time"

type RoomStatus string

const (
	RoomStatusOffline RoomStatus = "offline"
	RoomStatusLive    RoomStatus = "live"
)

type Room struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	OwnerUserID string     `json:"ownerUserId"`
	Slug        string     `gorm:"uniqueIndex" json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StreamKey   string     `gorm:"uniqueIndex;column:stream_key" json:"streamKey"`
	Status      RoomStatus `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (Room) TableName() string { return "rooms" }

type CreateRoomInput struct {
	OwnerUserID string
	Title       string
	Description string
}
