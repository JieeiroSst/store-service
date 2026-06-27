package history

import "time"

type Message struct {
	ID        int64
	UserID    string
	Role      string // "user" | "assistant"
	Content   string
	Source    string // "ollama"; empty for role "user"
	CreatedAt time.Time
}
