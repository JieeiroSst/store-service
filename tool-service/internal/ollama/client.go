package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL string
	mu      sync.RWMutex
	model   string
	http    *http.Client
}

func NewClient(baseURL, model string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Model() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.model
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Format   string    `json:"format,omitempty"`
	Options  any       `json:"options,omitempty"`
}

type chatResponse struct {
	Message Message `json:"message"`
	Error   string  `json:"error,omitempty"`
}

func (c *Client) Chat(ctx context.Context, system, user string, jsonMode bool) (string, error) {
	return c.ChatWith(ctx, c.Model(), system, user, jsonMode)
}

func (c *Client) ChatWith(ctx context.Context, model, system, user string, jsonMode bool) (string, error) {
	return c.chat(ctx, model, system, user, jsonMode, 0.2)
}

func (c *Client) chat(ctx context.Context, model, system, user string, jsonMode bool, temp float64) (string, error) {
	req := chatRequest{
		Model: model,
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Options: map[string]any{"temperature": temp},
	}
	if jsonMode {
		req.Format = "json"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call ollama: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode ollama response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || out.Error != "" {
		return "", fmt.Errorf("ollama status %d: %s", resp.StatusCode, out.Error)
	}
	return out.Message.Content, nil
}

func (c *Client) ChatJSON(ctx context.Context, system, user string, v any) error {
	return c.ChatJSONWith(ctx, c.Model(), system, user, v)
}

func (c *Client) ChatJSONWith(ctx context.Context, model, system, user string, v any) error {
	text, err := c.ChatWith(ctx, model, system, user, true)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(ExtractJSON(text)), v)
}

func (c *Client) ChatJSONTemp(ctx context.Context, system, user string, temp float64, v any) error {
	text, err := c.chat(ctx, c.Model(), system, user, true, temp)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(ExtractJSON(text)), v)
}

func ExtractJSON(s string) string {
	s = strings.TrimSpace(s)
	start := strings.IndexAny(s, "{[")
	if start < 0 {
		return s
	}
	end := strings.LastIndexAny(s, "}]")
	if end < start {
		return s
	}
	return s[start : end+1]
}

var ErrEmpty = errors.New("empty response from model")

func (c *Client) SetModel(m string) {
	c.mu.Lock()
	c.model = m
	c.mu.Unlock()
}

func (c *Client) Embed(ctx context.Context, model, text string) ([]float32, error) {
	body, _ := json.Marshal(map[string]any{"model": model, "input": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ollama embed: %w", err)
	}
	defer resp.Body.Close()
	var out struct {
		Embeddings [][]float32 `json:"embeddings"`
		Error      string      `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" || len(out.Embeddings) == 0 {
		return nil, fmt.Errorf("ollama embed: %s", out.Error)
	}
	return out.Embeddings[0], nil
}

func (c *Client) Create(ctx context.Context, name, base, system string, msgs []Message) error {
	body, _ := json.Marshal(map[string]any{
		"model": name, "from": base, "system": system, "messages": msgs, "stream": false,
		"parameters": map[string]any{"temperature": 0.2},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/create", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call ollama create: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama create status %d: %s", resp.StatusCode, raw)
	}
	return nil
}

func (c *Client) Pull(ctx context.Context, model string) error {
	body, _ := json.Marshal(map[string]any{"model": model, "stream": false})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call ollama pull: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama pull status %d: %s", resp.StatusCode, raw)
	}
	return nil
}
