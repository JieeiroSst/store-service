package model

import "time"

type MessageType string

const (
	TEXT  MessageType = "text"
	IMAGE MessageType = "image"
	VIDEO MessageType = "video"
	AUDIO MessageType = "audio"
)

type Type string

const (
	SINGLE Type = "single"
	GROUP  Type = "group"
)

type STATUS string

const (
	PENDING  STATUS = "pending"
	RESOLVED STATUS = "resolved"
)

type Messages struct {
	ID          int         `json:"_id,omitempty" bson:"_id,omitempty"`
	Guid        string      `json:"guid,omitempty" bson:"guid,omitempty"`
	SenderId    int         `json:"sender_id,omitempty" bson:"sender_id,omitempty"`
	MessageType MessageType `json:"message_type,omitempty" bson:"message_type,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt   time.Time   `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

type Participants struct {
	ID             int          `json:"_id,omitempty" bson:"_id,omitempty"`
	ConversationId int          `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	UsersId        int          `json:"users_id,omitempty" bson:"users_id,omitempty"`
	Type           Type         `json:"type,omitempty" bson:"type,omitempty"`
	CreatedAt      time.Time    `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt      time.Time    `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Conversation   Conversation `json:"conversation,omitempty" bson:"conversation,omitempty"`
}

type Reports struct {
	ID             int       `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId         int       `json:"users_id,omitempty" bson:"users_id,omitempty"`
	ParticipantsId int       `json:"participants_id,omitempty" bson:"participants_id,omitempty"`
	ReportType     string    `json:"report_type,omitempty" bson:"report_type,omitempty"`
	Notes          string    `json:"notes,omitempty" bson:"notes,omitempty"`
	Status         STATUS    `json:"status,omitempty" bson:"status,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

type BlockList struct {
	ID         int       `json:"_id,omitempty" bson:"_id,omitempty"`
	MessagesId int       `json:"messages_id,omitempty" bson:"messages_id,omitempty"`
	UsersId    int       `json:"users_id,omitempty" bson:"users_id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt  time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Messages   Messages  `json:"messages,omitempty" bson:"messages,omitempty"`
}

type DeletedMessages struct {
	ID         int       `json:"_id,omitempty" bson:"_id,omitempty"`
	MessagesId int       `json:"messages_id,omitempty" bson:"messages_id,omitempty"`
	UserId     int       `json:"users_id,omitempty" bson:"users_id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt  time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Messages   Messages  `json:"messages,omitempty" bson:"messages,omitempty"`
}

type Conversation struct {
	ID        int       `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     string    `json:"title,omitempty" bson:"title,omitempty"`
	CreatorId int       `json:"creator_id,omitempty" bson:"creator_id,omitempty"`
	ChannelId int       `json:"channel_id,omitempty" bson:"channel_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}
