package dto

import "time"

type Participants struct {
	ID             int
	ConversationId int
	UsersId        int
	Type           string
	CreatedAt      time.Time
	DeletedAt      time.Time
	Conversation   Conversation
}
