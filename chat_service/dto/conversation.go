package dto

import "time"

type Conversation struct {
	ID        int
	Title     string
	CreatorId int
	ChannelId int
	CreatedAt time.Time
	DeletedAt time.Time
}
