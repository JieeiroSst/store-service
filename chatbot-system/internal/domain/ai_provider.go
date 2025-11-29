package domain

import "context"

// AIProvider defines the strategy interface for different AI models
type AIProvider interface {
	SendMessage(ctx context.Context, conversation []Message, userMessage string) (string, error)
	GetModelName() string
	IsAvailable() bool
}

// AIProviderFactory creates AI providers based on model name
type AIProviderFactory interface {
	GetProvider(modelName string) (AIProvider, error)
	ListAvailableProviders() []string
}
