package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"chatbot-system/internal/core/domain"
)

type DeepSeekProvider struct {
	apiKey string
	model  string
}

// NewDeepSeekProvider creates a new DeepSeek AI provider
func NewDeepSeekProvider(apiKey, model string) *DeepSeekProvider {
	if model == "" {
		model = "deepseek-chat"
	}
	return &DeepSeekProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (d *DeepSeekProvider) GetModelName() string {
	return "deepseek"
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []deepSeekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

type deepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (d *DeepSeekProvider) SendMessage(ctx context.Context, history []domain.Message, userMessage string) (string, error) {
	// Build message history
	var messages []deepSeekMessage
	for _, msg := range history {
		messages = append(messages, deepSeekMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current message if not already in history
	if len(messages) == 0 || messages[len(messages)-1].Content != userMessage {
		messages = append(messages, deepSeekMessage{
			Role:    "user",
			Content: userMessage,
		})
	}

	requestBody := deepSeekRequest{
		Model:    d.model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deepSeekResp deepSeekResponse
	if err := json.Unmarshal(body, &deepSeekResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(deepSeekResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from DeepSeek")
	}

	return deepSeekResp.Choices[0].Message.Content, nil
}
