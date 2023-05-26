package dto

import "time"

type DeletedMessages struct {
	ID         int
	MessagesId int
	UserId     int
	CreatedAt  time.Time
	DeletedAt  time.Time
	Messages   Messages
}
