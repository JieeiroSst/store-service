package proxy

import "time"

type Config struct {
	OllamaBaseURL  string
	OllamaModel    string
	RequestTimeout time.Duration
	MaxRetries     int
}

type Client struct {
	Ollama OllamaClient
}

func New(cfg Config) *Client {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}

	return &Client{
		Ollama: newOllamaClient(cfg.OllamaBaseURL, cfg.OllamaModel, timeout, maxRetries),
	}
}
