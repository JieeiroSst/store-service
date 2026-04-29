package ollama

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"` 
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  Options   `json:"options,omitempty"`
}

type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumCtx      int     `json:"num_ctx,omitempty"`
}

// ChatResponse chunk from streaming
type ChatResponse struct {
	Model     string  `json:"model"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
	Error     string  `json:"error,omitempty"`
}

type StreamChunk struct {
	Type    string `json:"type"`    // "token" | "done" | "error"
	Content string `json:"content"` // text token or error message
	Source  string `json:"source"`  // "preset" | "ollama" | "external_api"
}

func (c *Client) StreamChat(
	history []Message,
	userMsg string,
	imagePaths []string,
	onChunk func(StreamChunk),
) error {
	msg := Message{
		Role:    "user",
		Content: userMsg,
	}

	for _, imgPath := range imagePaths {
		b64, err := encodeImage(imgPath)
		if err != nil {
			return fmt.Errorf("encode image %s: %w", imgPath, err)
		}
		msg.Images = append(msg.Images, b64)
	}

	messages := append(history, msg)

	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
		Options: Options{
			Temperature: 0.7,
			NumCtx:      4096,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/chat",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var chunk ChatResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}

		if chunk.Error != "" {
			onChunk(StreamChunk{Type: "error", Content: chunk.Error})
			return fmt.Errorf("ollama: %s", chunk.Error)
		}

		if chunk.Message.Content != "" {
			onChunk(StreamChunk{
				Type:    "token",
				Content: chunk.Message.Content,
				Source:  "ollama",
			})
		}

		if chunk.Done {
			onChunk(StreamChunk{Type: "done", Source: "ollama"})
			break
		}
	}

	return scanner.Err()
}

func encodeImage(pathOrBase64 string) (string, error) {
	if !strings.Contains(pathOrBase64, "/") && !strings.Contains(pathOrBase64, "\\") {
		return pathOrBase64, nil
	}

	data, err := os.ReadFile(pathOrBase64)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
