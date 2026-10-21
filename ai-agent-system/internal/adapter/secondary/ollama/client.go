package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aiagentsystem/internal/domain/model"
	"aiagentsystem/internal/domain/port"
)

const (
	defaultTimeout        = 60 * time.Second
	defaultMaxRetries     = 2
	defaultRetryBaseDelay = 500 * time.Millisecond
)

type Config struct {
	BaseURL        string
	Model          string
	Timeout        time.Duration
	MaxRetries     int
	RetryBaseDelay time.Duration
}

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
	retry      retryPolicy
}

var _ port.LLMClient = (*Client)(nil)

func New(cfg Config) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = defaultMaxRetries
	}
	if cfg.RetryBaseDelay <= 0 {
		cfg.RetryBaseDelay = defaultRetryBaseDelay
	}
	return &Client{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		model:      cfg.Model,
		httpClient: &http.Client{Timeout: cfg.Timeout},
		retry:      retryPolicy{maxRetries: cfg.MaxRetries, baseDelay: cfg.RetryBaseDelay},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	UserID   string        `json:"user_id,omitempty"`
}

type chatChunk struct {
	Message         chatMessage `json:"message"`
	Done            bool        `json:"done"`
	DoneReason      string      `json:"done_reason,omitempty"`
	PromptEvalCount int64       `json:"prompt_eval_count"`
	EvalCount       int64       `json:"eval_count"`
}

func (c *Client) Complete(ctx context.Context, q model.Question) (model.Completion, error) {
	resp, err := c.post(ctx, q, false)
	if err != nil {
		return model.Completion{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Completion{}, internalError("failed to read Ollama response", err)
	}
	if resp.StatusCode != http.StatusOK {
		return model.Completion{}, statusError(resp.StatusCode, string(body))
	}

	var chunk chatChunk
	if err := json.Unmarshal(body, &chunk); err != nil {
		return model.Completion{}, internalError("failed to parse Ollama response", fmt.Errorf("%w (body: %s)", err, body))
	}

	return model.Completion{
		Text:         chunk.Message.Content,
		StopReason:   chunk.DoneReason,
		InputTokens:  chunk.PromptEvalCount,
		OutputTokens: chunk.EvalCount,
	}, nil
}

func (c *Client) Stream(ctx context.Context, q model.Question) (port.DeltaStream, error) {
	resp, err := c.post(ctx, q, true)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, statusError(resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &deltaStream{body: resp.Body, scanner: scanner}, nil
}

func (c *Client) post(ctx context.Context, q model.Question, stream bool) (*http.Response, error) {
	messages := make([]chatMessage, 0, len(q.Messages))
	for _, m := range q.Messages {
		messages = append(messages, chatMessage{Role: m.Role, Content: m.Content})
	}
	body, err := json.Marshal(chatRequest{Model: c.model, Messages: messages, Stream: stream, UserID: q.UserID})
	if err != nil {
		return nil, model.NewError(model.KindBadRequest, "failed to build Ollama request", err)
	}

	resp, err := c.retry.do(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return c.httpClient.Do(req)
	})
	if err != nil {
		return nil, unreachableError(err)
	}
	return resp, nil
}

type deltaStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	current string
	text    strings.Builder
	final   model.Completion
	err     error
}

func (s *deltaStream) Next() bool {
	for s.scanner.Scan() {
		line := bytes.TrimSpace(s.scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var chunk chatChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			s.err = internalError("failed to parse Ollama stream chunk", err)
			return false
		}

		if chunk.Done {
			s.final = model.Completion{
				Text:         s.text.String(),
				StopReason:   chunk.DoneReason,
				InputTokens:  chunk.PromptEvalCount,
				OutputTokens: chunk.EvalCount,
			}
			return false
		}
		if chunk.Message.Content != "" {
			s.text.WriteString(chunk.Message.Content)
			s.current = chunk.Message.Content
			return true
		}
	}
	if err := s.scanner.Err(); err != nil {
		s.err = unreachableError(err)
	}
	return false
}

func (s *deltaStream) Current() string          { return s.current }
func (s *deltaStream) Result() model.Completion { return s.final }
func (s *deltaStream) Err() error               { return s.err }
func (s *deltaStream) Close() error             { return s.body.Close() }
