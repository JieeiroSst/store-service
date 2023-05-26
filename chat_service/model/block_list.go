package model

import "time"

type BlockList struct {
	ID         int       `json:"_id,omitempty" bson:"_id,omitempty"`
	MessagesId int       `json:"messages_id,omitempty" bson:"messages_id,omitempty"`
	UsersId    int       `json:"users_id,omitempty" bson:"users_id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	DeletedAt  time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Messages   Messages  `json:"messages,omitempty" bson:"messages,omitempty"`
}
