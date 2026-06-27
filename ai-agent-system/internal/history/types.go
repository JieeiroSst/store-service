package history

import "time"

type Message struct {
	ID        int64
	UserID    string
	Role      string // "user" | "assistant"
	Content   string
	Source    string // "faq" | "claude"; empty for role "user"
	CreatedAt time.Time
}
