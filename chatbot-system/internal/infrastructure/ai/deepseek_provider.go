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

type DeepSeekProvider struct {
	apiKey     string
	httpClient *http.Client
	modelName  string
}

func NewDeepSeekProvider(apiKey string) *DeepSeekProvider {
	return &DeepSeekProvider{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		modelName:  "deepseek-chat",
	}
}

type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekRequest struct {
	Model    string            `json:"model"`
	Messages []deepseekMessage `json:"messages"`
}

type deepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (d *DeepSeekProvider) SendMessage(ctx context.Context, conversation []domain.Message, userMessage string) (string, error) {
	messages := make([]deepseekMessage, 0, len(conversation)+1)
	
	// Convert conversation history
	for _, msg := range conversation {
		role := "user"
		if msg.MessageType == domain.MessageTypeAI {
			role = "assistant"
		}
		messages = append(messages, deepseekMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	
	// Add current message
	messages = append(messages, deepseekMessage{
		Role:    "user",
		Content: userMessage,
	})

	reqBody := deepseekRequest{
		Model:    d.modelName,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var deepseekResp deepseekResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return deepseekResp.Choices[0].Message.Content, nil
}

func (d *DeepSeekProvider) GetModelName() string {
	return "deepseek"
}

func (d *DeepSeekProvider) IsAvailable() bool {
	return d.apiKey != ""
}
