package model

import "time"

type MessageType string

const (
	TEXT  MessageType = "text"
	IMAGE MessageType = "image"
	VIDEO MessageType = "video"
	AUDIO MessageType = "audio"
)

type Messages struct {
	ID          int         `json:"_id,omitempty" bson:"_id,omitempty"`
	Guid        string      `json:"guid,omitempty" bson:"guid,omitempty"`
	SenderId    int         `json:"sender_id,omitempty" bson:"sender_id,omitempty"`
	MessageType MessageType `json:"message_type,omitempty" bson:"message_type,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt   time.Time   `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}
