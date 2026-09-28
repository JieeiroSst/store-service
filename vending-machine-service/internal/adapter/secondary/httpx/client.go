package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxErrorBody = 4 << 10

type Client struct {
	base        string
	http        *http.Client
	requestedBy string
}

func New(baseURL string, timeout time.Duration, requestedBy string) *Client {
	return &Client{
		base:        strings.TrimRight(baseURL, "/"),
		http:        &http.Client{Timeout: timeout},
		requestedBy: requestedBy,
	}
}

type StatusError struct {
	Status  int
	Message string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("status %d: %s", e.Status, e.Message)
}

func IsStatus(err error, codes ...int) bool {
	var se *StatusError
	if !errors.As(err, &se) {
		return false
	}
	for _, c := range codes {
		if se.Status == c {
			return true
		}
	}
	return false
}

func Retryable(err error) bool {
	var se *StatusError
	if errors.As(err, &se) {
		return se.Status >= 500
	}
	return !errors.Is(err, context.Canceled)
}

func (c *Client) Do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.requestedBy != "" {
		req.Header.Set("X-Requested-By", c.requestedBy)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		var payload struct {
			Error string `json:"error"`
		}
		msg := strings.TrimSpace(string(raw))
		if json.Unmarshal(raw, &payload) == nil && payload.Error != "" {
			msg = payload.Error
		}
		return &StatusError{Status: resp.StatusCode, Message: msg}
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s %s response: %w", method, path, err)
	}
	return nil
}

func WithRetry(ctx context.Context, attempts int, backoff time.Duration, fn func() error) error {
	var err error
	for i := range attempts {
		if err = fn(); err == nil || !Retryable(err) {
			return err
		}
		if i == attempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return errors.Join(err, ctx.Err())
		case <-time.After(backoff * time.Duration(i+1)):
		}
	}
	return err
}
