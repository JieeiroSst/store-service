package model

import "time"

type Turn struct {
	UserID   string
	Question string
	Answer   string
	Source   Source
}

type HistoryEntry struct {
	ID        int64
	UserID    string
	Role      string
	Content   string
	Source    string
	CreatedAt time.Time
}
