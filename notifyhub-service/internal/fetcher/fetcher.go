package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
)

type Fetcher struct {
	client *http.Client
}

func New(poolSize, timeoutSec int) *Fetcher {
	transport := &http.Transport{
		MaxIdleConns:          poolSize,
		MaxIdleConnsPerHost:   poolSize / 2,
		MaxConnsPerHost:       poolSize,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false,
	}
	return &Fetcher{
		client: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(timeoutSec) * time.Second,
		},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, ds *model.DataSource) (interface{}, error) {
	var bodyReader io.Reader
	if ds.Body != "" {
		bodyReader = bytes.NewBufferString(ds.Body)
	}

	req, err := http.NewRequestWithContext(ctx, ds.Method, ds.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request for %q: %w", ds.Name, err)
	}

	if ds.Method == "POST" || ds.Method == "PUT" {
		if _, ok := ds.Headers["Content-Type"]; !ok {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	req.Header.Set("Accept", "application/json")

	for k, v := range ds.Headers {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}

	if err := applyAuth(req, ds.AuthType, ds.AuthConfig); err != nil {
		return nil, fmt.Errorf("apply auth: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request to %q: %w", ds.URL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB max
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d from %q: %s", resp.StatusCode, ds.URL, truncate(string(raw), 200))
	}

	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return strings.TrimSpace(string(raw)), nil
	}

	if ds.JSONPath != "" {
		parsed = extractPath(parsed, ds.JSONPath)
	}

	return parsed, nil
}

func (f *Fetcher) FetchMany(ctx context.Context, sources []*model.DataSource) map[string]interface{} {
	results := make(map[string]interface{}, len(sources))
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for _, ds := range sources {
		wg.Add(1)
		go func(ds *model.DataSource) {
			defer wg.Done()
			data, err := f.Fetch(ctx, ds)
			mu.Lock()
			if err == nil {
				results[ds.Name] = data
			} else {
				results[ds.Name] = nil
			}
			mu.Unlock()
		}(ds)
	}
	wg.Wait()
	return results
}

func applyAuth(req *http.Request, authType string, cfg model.JSONMap) error {
	switch strings.ToLower(authType) {
	case "bearer":
		token, _ := cfg["token"].(string)
		if token == "" {
			return fmt.Errorf("bearer auth: token is empty")
		}
		req.Header.Set("Authorization", "Bearer "+token)

	case "basic":
		user, _ := cfg["username"].(string)
		pass, _ := cfg["password"].(string)
		req.SetBasicAuth(user, pass)

	case "apikey":
		key, _ := cfg["key"].(string)
		header, _ := cfg["header"].(string)
		if header == "" {
			header = "X-API-Key"
		}
		param, _ := cfg["param"].(string)
		if param != "" {
			q := req.URL.Query()
			q.Set(param, key)
			req.URL.RawQuery = q.Encode()
		} else {
			req.Header.Set(header, key)
		}

	case "none", "":
		// No auth required
	default:
		return fmt.Errorf("unknown auth_type %q", authType)
	}
	return nil
}

func extractPath(data interface{}, path string) interface{} {
	// Strip leading "$." or "$"
	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, "$")
	if path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if part == "" {
			continue
		}
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			// Try to parse as integer index
			var idx int
			if _, err := fmt.Sscanf(part, "[%d]", &idx); err == nil {
				if idx >= 0 && idx < len(v) {
					current = v[idx]
				} else {
					return nil
				}
			} else {
				return current
			}
		default:
			return current
		}
	}
	return current
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
