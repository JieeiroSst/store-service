package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EmbeddingClient interface {
	Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error)
}

type ollamaEmbedClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
	maxRetries int
}

func newOllamaEmbedClient(baseURL, model string, timeout time.Duration, maxRetries int) EmbeddingClient {
	return &ollamaEmbedClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
		maxRetries: maxRetries,
	}
}

type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed calls Ollama's /api/embed. inputType (Voyage's "document" vs "query"
// distinction) has no Ollama equivalent and is ignored.
func (c *ollamaEmbedClient) Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody, err := json.Marshal(ollamaEmbedRequest{
		Model: c.model,
		Input: texts,
	})
	if err != nil {
		return nil, &Error{Kind: ErrKindInvalidRequest, Message: "failed to build embeddings request", Err: err}
	}

	resp, err := doWithRetry(ctx, c.maxRetries, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return c.httpClient.Do(req)
	})
	if err != nil {
		return nil, mapOllamaError(err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, &Error{Kind: ErrKindUnknown, Message: "failed to read embeddings response", Err: readErr}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, mapOllamaHTTPError(resp.StatusCode, string(body))
	}

	var parsed ollamaEmbedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, &Error{Kind: ErrKindUnknown, Message: "failed to parse embeddings response", Err: fmt.Errorf("%w (body: %s)", err, body)}
	}

	return parsed.Embeddings, nil
}
