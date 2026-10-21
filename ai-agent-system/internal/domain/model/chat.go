package model

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"

	AnonymousUserID = "anonymous"
)

type Message struct {
	Role    string
	Content string
}

type Question struct {
	Messages []Message
	UserID   string
}

func (q Question) LastUserQuestion() string {
	if len(q.Messages) == 0 {
		return ""
	}
	return q.Messages[len(q.Messages)-1].Content
}

type Source string

const SourceOllama Source = "ollama"

type Completion struct {
	Text         string
	StopReason   string
	InputTokens  int64
	OutputTokens int64
}

type Answer struct {
	Completion
	Source Source
}
