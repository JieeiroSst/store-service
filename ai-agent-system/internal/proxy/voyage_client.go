package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const voyageModel = "voyage-4-large"

const voyageEmbeddingsURL = "https://api.voyageai.com/v1/embeddings"

type VoyageClient interface {
	Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error)
}

type voyageClient struct {
	apiKey     string
	httpClient *http.Client
	maxRetries int
}

func newVoyageClient(apiKey string, timeout time.Duration, maxRetries int) VoyageClient {
	return &voyageClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: timeout},
		maxRetries: maxRetries,
	}
}

type voyageEmbedRequest struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type,omitempty"`
}

type voyageEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *voyageClient) Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody, err := json.Marshal(voyageEmbedRequest{
		Input:     texts,
		Model:     voyageModel,
		InputType: inputType,
	})
	if err != nil {
		return nil, &Error{Kind: ErrKindInvalidRequest, Message: "failed to build embeddings request", Err: err}
	}

	resp, err := doWithRetry(ctx, c.maxRetries, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, voyageEmbeddingsURL, bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		return c.httpClient.Do(req)
	})
	if err != nil {
		return nil, &Error{Kind: ErrKindUnknown, Message: "failed to reach the embeddings provider", Err: err}
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, &Error{Kind: ErrKindUnknown, Message: "failed to read embeddings response", Err: readErr}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, mapVoyageHTTPError(resp.StatusCode, string(body))
	}

	var parsed voyageEmbedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, &Error{Kind: ErrKindUnknown, Message: "failed to parse embeddings response", Err: fmt.Errorf("%w (body: %s)", err, body)}
	}

	embeddings := make([][]float32, len(texts))
	for _, item := range parsed.Data {
		if item.Index < 0 || item.Index >= len(embeddings) {
			continue
		}
		embeddings[item.Index] = item.Embedding
	}
	return embeddings, nil
}
