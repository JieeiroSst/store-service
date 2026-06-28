package proxy

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
)

type OllamaClient interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResult, error)
	Stream(ctx context.Context, req CompletionRequest) (DeltaStream, error)
}

type ollamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
	maxRetries int
}

func newOllamaClient(baseURL, model string, timeout time.Duration, maxRetries int) OllamaClient {
	return &ollamaClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
		maxRetries: maxRetries,
	}
}

type ollamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	UserID   string              `json:"user_id,omitempty"`
}

type ollamaChatChunk struct {
	Message         ollamaChatMessage `json:"message"`
	Done            bool              `json:"done"`
	DoneReason      string            `json:"done_reason,omitempty"`
	PromptEvalCount int64             `json:"prompt_eval_count"`
	EvalCount       int64             `json:"eval_count"`
}

func buildOllamaMessages(req CompletionRequest) []ollamaChatMessage {
	messages := make([]ollamaChatMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, ollamaChatMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		messages = append(messages, ollamaChatMessage{Role: m.Role, Content: m.Content})
	}
	return messages
}

func (c *ollamaClient) doChatRequest(ctx context.Context, stream bool, req CompletionRequest) (*http.Response, error) {
	body, err := json.Marshal(ollamaChatRequest{
		Model:    c.model,
		Messages: buildOllamaMessages(req),
		Stream:   stream,
		UserID:   req.UserID,
	})
	if err != nil {
		return nil, &Error{Kind: ErrKindInvalidRequest, Message: "failed to build Ollama request", Err: err}
	}

	resp, err := doWithRetry(ctx, c.maxRetries, func() (*http.Response, error) {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		return c.httpClient.Do(httpReq)
	})
	if err != nil {
		return nil, mapOllamaError(err)
	}
	return resp, nil
}

func (c *ollamaClient) Complete(ctx context.Context, req CompletionRequest) (CompletionResult, error) {
	resp, err := c.doChatRequest(ctx, false, req)
	if err != nil {
		return CompletionResult{}, err
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return CompletionResult{}, &Error{Kind: ErrKindUnknown, Message: "failed to read Ollama response", Err: readErr}
	}
	if resp.StatusCode != http.StatusOK {
		return CompletionResult{}, mapOllamaHTTPError(resp.StatusCode, string(respBody))
	}

	var chunk ollamaChatChunk
	if err := json.Unmarshal(respBody, &chunk); err != nil {
		return CompletionResult{}, &Error{Kind: ErrKindUnknown, Message: "failed to parse Ollama response", Err: fmt.Errorf("%w (body: %s)", err, respBody)}
	}

	return CompletionResult{
		Text:         chunk.Message.Content,
		StopReason:   chunk.DoneReason,
		InputTokens:  chunk.PromptEvalCount,
		OutputTokens: chunk.EvalCount,
	}, nil
}

func (c *ollamaClient) Stream(ctx context.Context, req CompletionRequest) (DeltaStream, error) {
	resp, err := c.doChatRequest(ctx, true, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, mapOllamaHTTPError(resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &ollamaDeltaStream{body: resp.Body, scanner: scanner}, nil
}

type ollamaDeltaStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	current string
	final   CompletionResult
	text    strings.Builder
	err     error
}

func (s *ollamaDeltaStream) Next() bool {
	for s.scanner.Scan() {
		line := bytes.TrimSpace(s.scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var chunk ollamaChatChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			s.err = &Error{Kind: ErrKindUnknown, Message: "failed to parse Ollama stream chunk", Err: err}
			return false
		}

		if chunk.Done {
			s.final = CompletionResult{
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
		// Empty, non-final chunk — keep reading.
	}
	if err := s.scanner.Err(); err != nil {
		s.err = mapOllamaError(err)
	}
	return false
}

func (s *ollamaDeltaStream) Current() string { return s.current }

func (s *ollamaDeltaStream) Result() CompletionResult { return s.final }

func (s *ollamaDeltaStream) Err() error { return s.err }

func (s *ollamaDeltaStream) Close() error { return s.body.Close() }
