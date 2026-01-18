package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	CircleAPIBaseURL = "https://api.circle.com"
	CircleAPIVersion = "v1"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: CircleAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Response struct {
	Data interface{} `json:"data"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) Do(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	fullURL := fmt.Sprintf("%s/%s%s", c.baseURL, CircleAPIVersion, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
			return fmt.Errorf("circle API error: %d - %s", resp.StatusCode, string(bodyBytes))
		}
		return fmt.Errorf("circle API error %d: %s", errResp.Code, errResp.Message)
	}

	if result != nil {
		var apiResp struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

func BuildQueryParams(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	query := url.Values{}
	for k, v := range params {
		if v != "" {
			query.Add(k, v)
		}
	}

	if len(query) == 0 {
		return ""
	}

	return "?" + query.Encode()
}
