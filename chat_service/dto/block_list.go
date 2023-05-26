package dto

import "time"

type BlockList struct {
	ID         int
	MessagesId int
	UsersId    int
	CreatedAt  time.Time
	DeletedAt  time.Time
	Messages   Messages
}
