package model

import "time"

type ChatMessage struct {
	RoomID   string    `json:"roomId"`
	UserID   string    `json:"userId"`
	Username string    `json:"username"`
	Body     string    `json:"body"`
	SentAt   time.Time `json:"sentAt"`
}
