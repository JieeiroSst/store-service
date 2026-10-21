package ollama

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const maxRetryDelay = 10 * time.Second

type retryPolicy struct {
	maxRetries int
	baseDelay  time.Duration
}

func (p retryPolicy) do(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		resp, err := fn()
		if err == nil && !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		lastResp, lastErr = resp, err
		if attempt == p.maxRetries {
			break
		}
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		delay := p.baseDelay * time.Duration(1<<attempt)
		if delay > maxRetryDelay {
			delay = maxRetryDelay
		}
		delay += time.Duration(rand.Int63n(int64(p.baseDelay)))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastResp, lastErr
}

func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}
