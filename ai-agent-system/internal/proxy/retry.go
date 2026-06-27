package proxy

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func doWithRetry(ctx context.Context, maxRetries int, fn func() (*http.Response, error)) (*http.Response, error) {
	const (
		baseDelay = 500 * time.Millisecond
		maxDelay  = 10 * time.Second
	)

	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := fn()
		if err == nil && !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		if err == nil && isRetryableStatus(resp.StatusCode) && attempt < maxRetries {
			// Drain and close the body before retrying so the connection can be reused.
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		lastResp, lastErr = resp, err
		if attempt == maxRetries {
			break
		}

		delay := baseDelay * time.Duration(1<<attempt)
		if delay > maxDelay {
			delay = maxDelay
		}
		delay += time.Duration(rand.Int63n(int64(baseDelay)))

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
