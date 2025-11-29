package ai

import (
	"fmt"

	"chatbot-system/internal/domain"
)

type ProviderFactory struct {
	providers map[string]domain.AIProvider
}

func NewProviderFactory(claudeAPIKey, deepseekAPIKey string) *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[string]domain.AIProvider),
	}

	if claudeAPIKey != "" {
		factory.providers["claude"] = NewClaudeProvider(claudeAPIKey)
	}

	if deepseekAPIKey != "" {
		factory.providers["deepseek"] = NewDeepSeekProvider(deepseekAPIKey)
	}

	return factory
}

func (f *ProviderFactory) GetProvider(modelName string) (domain.AIProvider, error) {
	provider, exists := f.providers[modelName]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", modelName)
	}

	if !provider.IsAvailable() {
		return nil, fmt.Errorf("provider '%s' is not available", modelName)
	}

	return provider, nil
}

func (f *ProviderFactory) ListAvailableProviders() []string {
	var available []string
	for name, provider := range f.providers {
		if provider.IsAvailable() {
			available = append(available, name)
		}
	}
	return available
}
