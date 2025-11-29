package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"chatbot-system/internal/domain"
)

type ClaudeProvider struct {
	apiKey     string
	httpClient *http.Client
	modelName  string
}

func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		modelName:  "claude-sonnet-4-20250514",
	}
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model      string          `json:"model"`
	MaxTokens  int             `json:"max_tokens"`
	Messages   []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (c *ClaudeProvider) SendMessage(ctx context.Context, conversation []domain.Message, userMessage string) (string, error) {
	messages := make([]claudeMessage, 0, len(conversation)+1)
	
	// Convert conversation history
	for _, msg := range conversation {
		role := "user"
		if msg.MessageType == domain.MessageTypeAI {
			role = "assistant"
		}
		messages = append(messages, claudeMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	
	// Add current message
	messages = append(messages, claudeMessage{
		Role:    "user",
		Content: userMessage,
	})

	reqBody := claudeRequest{
		Model:     c.modelName,
		MaxTokens: 4096,
		Messages:  messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return claudeResp.Content[0].Text, nil
}

func (c *ClaudeProvider) GetModelName() string {
	return "claude"
}

func (c *ClaudeProvider) IsAvailable() bool {
	return c.apiKey != ""
}
