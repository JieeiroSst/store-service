package model

import "time"

type Message struct {
	ID        uint      `json:"id"`
	RoomID    uint      `json:"room_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type EventType string

const (
	EventMessage       EventType = "message"
	EventMemberAdded   EventType = "member_added"
	EventMemberRemoved EventType = "member_removed"
)

type Event struct {
	Type    EventType `json:"type"`
	RoomID  uint      `json:"room_id"`
	Message *Message  `json:"message,omitempty"`
	Member  *Member   `json:"member,omitempty"`
}
