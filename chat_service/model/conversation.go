package model

import "time"

type Conversation struct {
	ID        int       `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     string    `json:"title,omitempty" bson:"title,omitempty"`
	CreatorId int       `json:"creator_id,omitempty" bson:"creator_id,omitempty"`
	ChannelId int       `json:"channel_id,omitempty" bson:"channel_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}
