package proxy

type Message struct {
	Role    string // "user" | "assistant"
	Content string
}

type CompletionRequest struct {
	SystemPrompt string
	Messages     []Message
	MaxTokens    int64
}

type CompletionResult struct {
	Text         string
	StopReason   string
	InputTokens  int64
	OutputTokens int64
}

type DeltaStream interface {
	Next() bool
	Current() string
	Result() CompletionResult
	Err() error
	Close() error
}
