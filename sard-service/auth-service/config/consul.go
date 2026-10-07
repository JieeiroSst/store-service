package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultConsulKey = "auth_service"
	consulAttempts   = 10
	consulBackoff    = 2 * time.Second
	consulTimeout    = 5 * time.Second
)

var errKeyMissing = errors.New("consul key not found")

func Load() (*Config, error) {
	cfg := Defaults()
	if addr := os.Getenv("CONSUL_ADDR"); addr != "" {
		key := os.Getenv("CONSUL_KEY")
		if key == "" {
			key = defaultConsulKey
		}
		raw, err := fetchConsul(context.Background(), addr, key)
		switch {
		case errors.Is(err, errKeyMissing):
			log.Printf("consul key %q not found, using defaults and environment", key)
		case err != nil:
			return nil, err
		default:
			if err := cfg.Merge(raw); err != nil {
				return nil, fmt.Errorf("consul key %q: %w", key, err)
			}
			log.Printf("loaded config from consul key %q", key)
		}
	}
	cfg.ApplyEnv()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func fetchConsul(ctx context.Context, addr, key string) ([]byte, error) {
	endpoint := strings.TrimRight(addr, "/") + "/v1/kv/" + url.PathEscape(key) + "?raw"
	client := &http.Client{Timeout: consulTimeout}
	var lastErr error
	for attempt := 1; attempt <= consulAttempts; attempt++ {
		raw, err := getKey(ctx, client, endpoint)
		if err == nil || errors.Is(err, errKeyMissing) {
			return raw, err
		}
		lastErr = err
		log.Printf("consul not ready (attempt %d/%d): %v", attempt, consulAttempts, err)
		time.Sleep(consulBackoff)
	}
	return nil, fmt.Errorf("read consul key %q: %w", key, lastErr)
}

func getKey(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if token := os.Getenv("CONSUL_HTTP_TOKEN"); token != "" {
		req.Header.Set("X-Consul-Token", token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, errKeyMissing
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("consul returned %s", resp.Status)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil, errKeyMissing
	}
	return raw, nil
}
