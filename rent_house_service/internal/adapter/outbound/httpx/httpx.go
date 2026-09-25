package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
	Name    string 
}

func New(name, baseURL string, timeout time.Duration) *Client {
	return &Client{HTTP: &http.Client{Timeout: timeout}, BaseURL: strings.TrimRight(baseURL, "/"), Name: name}
}

func (c *Client) Do(ctx context.Context, method, path string, headers map[string]string, in, out any) (int, string, error) {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return 0, "", err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return 0, "", err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("%w: %s: %v", domain.ErrUpstreamUnavailable, c.Name, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, "", fmt.Errorf("%w: %s: %v", domain.ErrUpstreamUnavailable, c.Name, err)
	}
	if resp.StatusCode >= 500 {
		return resp.StatusCode, "", fmt.Errorf("%w: %s returned %d", domain.ErrUpstreamUnavailable, c.Name, resp.StatusCode)
	}
	if resp.StatusCode/100 != 2 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &e)
		return resp.StatusCode, e.Error, nil
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return resp.StatusCode, "", fmt.Errorf("%w: %s: bad response: %v", domain.ErrUpstreamUnavailable, c.Name, err)
		}
	}
	return resp.StatusCode, "", nil
}

func OK(status int) bool { return status/100 == 2 }
