package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	RoomID    uint      `json:"room_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
