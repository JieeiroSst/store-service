package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	resendAPIURL      = "https://api.resend.com/emails"
	resendBatchAPIURL = "https://api.resend.com/emails/batch"
	MaxBatchSize      = 100
)

type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string { return fmt.Sprintf("resend api error (%d): %s", e.Status, e.Body) }

func (e *APIError) Retryable() bool { return e.Status == http.StatusTooManyRequests || e.Status >= 500 }

type Client struct {
	apiKey     string
	from       string
	baseURL    string
	batchURL   string
	httpClient *http.Client
}

func NewClient(apiKey, from string) *Client {
	return &Client{
		apiKey:   apiKey,
		from:     from,
		baseURL:  resendAPIURL,
		batchURL: resendBatchAPIURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (c *Client) Send(to []string, subject, body string) error {
	payload, err := json.Marshal(sendRequest{
		From:    c.from,
		To:      to,
		Subject: subject,
		HTML:    body,
	})
	if err != nil {
		return fmt.Errorf("marshal resend request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send resend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return &APIError{Status: resp.StatusCode, Body: string(respBody)}
	}
	return nil
}

type Message struct {
	To      string
	Subject string
	HTML    string
}

func (c *Client) WithEndpoints(sendURL, batchURL string) *Client {
	c.baseURL, c.batchURL = sendURL, batchURL
	return c
}

func (c *Client) SendBatch(ctx context.Context, messages []Message, idempotencyKey string) error {
	if len(messages) == 0 {
		return nil
	}
	if len(messages) > MaxBatchSize {
		return fmt.Errorf("resend batch accepts at most %d emails, got %d", MaxBatchSize, len(messages))
	}
	items := make([]sendRequest, len(messages))
	for i, m := range messages {
		items[i] = sendRequest{From: c.from, To: []string{m.To}, Subject: m.Subject, HTML: m.HTML}
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal resend batch: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.batchURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build resend batch request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send resend batch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return &APIError{Status: resp.StatusCode, Body: string(respBody)}
	}
	return nil
}
