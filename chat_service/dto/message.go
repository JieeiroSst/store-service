package dto

import "time"

type Messages struct {
	ID          int
	Guid        string
	SenderId    int
	MessageType string
	CreatedAt   time.Time
	DeletedAt   time.Time
}
