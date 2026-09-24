// Package httpx is the small JSON-over-HTTP client the sibling-service
// adapters share.
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

type Client struct {
	base string
	http *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{base: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: timeout}}
}

func (c *Client) Enabled() bool { return c.base != "" }

// ErrTransport marks failures that never produced an HTTP response.
var ErrTransport = errors.New("transport error")

// StatusError is a non-2xx answer from the service.
type StatusError struct {
	Status  int
	Message string
}

func (e *StatusError) Error() string { return fmt.Sprintf("status %d: %s", e.Status, e.Message) }

// Do sends body (if any) as JSON and decodes a 2xx response into out (if any).
func (c *Client) Do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTransport, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &StatusError{Status: resp.StatusCode, Message: message(raw)}
	}
	if out == nil || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// message digs the reason out of the error shapes our services use:
// {"error":"..."} (gin) and {"error":{"message":"..."}} (gofr).
// Message extracts the reason from a service error body.
func Message(raw []byte) string { return message(raw) }

func message(raw []byte) string {
	var p struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &p) == nil && len(p.Error) > 0 {
		var s string
		if json.Unmarshal(p.Error, &s) == nil && s != "" {
			return s
		}
		var o struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(p.Error, &o) == nil && o.Message != "" {
			return o.Message
		}
	}
	return strings.TrimSpace(string(raw))
}

// IsStatus reports whether err is a StatusError with the given code.
func IsStatus(err error, code int) bool {
	var se *StatusError
	return errors.As(err, &se) && se.Status == code
}

// AsStatus unwraps a StatusError from err.
func AsStatus(err error) (*StatusError, bool) {
	var se *StatusError
	ok := errors.As(err, &se)
	return se, ok
}
