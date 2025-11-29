package services

import (
	"fmt"
	"chatbot-system/internal/core/ports"
)

// AIProviderFactory manages AI provider strategies
type AIProviderFactory struct {
	providers map[string]ports.AIProvider
}

// NewAIProviderFactory creates a new factory
func NewAIProviderFactory() *AIProviderFactory {
	return &AIProviderFactory{
		providers: make(map[string]ports.AIProvider),
	}
}

// RegisterProvider registers an AI provider
func (f *AIProviderFactory) RegisterProvider(name string, provider ports.AIProvider) {
	f.providers[name] = provider
}

// GetProvider returns the appropriate AI provider based on the model name
func (f *AIProviderFactory) GetProvider(modelName string) (ports.AIProvider, error) {
	provider, exists := f.providers[modelName]
	if !exists {
		return nil, fmt.Errorf("AI provider '%s' not found", modelName)
	}
	return provider, nil
}

// GetAvailableProviders returns list of available provider names
func (f *AIProviderFactory) GetAvailableProviders() []string {
	var names []string
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}
