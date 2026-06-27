package chat

type Message struct {
	Role    string
	Content string
}

type AnswerRequest struct {
	Messages []Message
	UserID   string
}

func (r AnswerRequest) LastUserQuestion() string {
	if len(r.Messages) == 0 {
		return ""
	}
	return r.Messages[len(r.Messages)-1].Content
}

type Source string

const (
	SourceFAQ    Source = "faq"
	SourceOllama Source = "ollama"
)

type AnswerResult struct {
	Text         string
	Source       Source
	StopReason   string
	InputTokens  int64
	OutputTokens int64
}
