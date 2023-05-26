package model

import "time"

type Type string

const (
	SINGLE Type = "single"
	GROUP  Type = "group"
)

type Participants struct {
	ID             int          `json:"_id,omitempty" bson:"_id,omitempty"`
	ConversationId int          `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	UsersId        int          `json:"users_id,omitempty" bson:"users_id,omitempty"`
	Type           Type         `json:"type,omitempty" bson:"type,omitempty"`
	CreatedAt      time.Time    `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt      time.Time    `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Conversation   Conversation `json:"conversation,omitempty" bson:"conversation,omitempty"`
}
