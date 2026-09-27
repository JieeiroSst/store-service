package benchmark

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MaxConcurrency = 200
	MaxRequests    = 100000
	MaxDuration    = 5 * time.Minute
)

type Config struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	Concurrency int               `json:"concurrency"`

	Requests    int `json:"requests"`
	DurationSec int `json:"duration_sec"`
}

type Stats struct {
	URL           string         `json:"url"`
	Concurrency   int            `json:"concurrency"`
	TotalRequests int64          `json:"total_requests"`
	Success       int64          `json:"success"`
	Failed        int64          `json:"failed"`
	ElapsedMs     int64          `json:"elapsed_ms"`
	RPS           float64        `json:"rps"`
	MinMs         float64        `json:"min_ms"`
	MeanMs        float64        `json:"mean_ms"`
	P50Ms         float64        `json:"p50_ms"`
	P90Ms         float64        `json:"p90_ms"`
	P95Ms         float64        `json:"p95_ms"`
	P99Ms         float64        `json:"p99_ms"`
	MaxMs         float64        `json:"max_ms"`
	StatusCodes   map[int]int64  `json:"status_codes"`
	Errors        map[string]int `json:"errors,omitempty"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return errors.New("url is required")
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 10
	}
	if c.Concurrency > MaxConcurrency {
		return errors.New("concurrency exceeds limit")
	}
	if c.DurationSec <= 0 && c.Requests <= 0 {
		c.Requests = 100
	}
	if c.Requests > MaxRequests {
		return errors.New("requests exceeds limit")
	}
	if time.Duration(c.DurationSec)*time.Second > MaxDuration {
		return errors.New("duration exceeds limit")
	}
	if c.Method == "" {
		c.Method = http.MethodGet
	}
	c.Method = strings.ToUpper(c.Method)
	return nil
}

func Run(ctx context.Context, client *http.Client, cfg Config) Stats {
	if cfg.DurationSec > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(cfg.DurationSec)*time.Second)
		defer cancel()
	}

	var (
		mu        sync.Mutex
		latencies []float64
		codes     = map[int]int64{}
		errs      = map[string]int{}
		success   int64
		failed    int64
		issued    int64
		wg        sync.WaitGroup
	)

	start := time.Now()
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				if cfg.DurationSec <= 0 && atomic.AddInt64(&issued, 1) > int64(cfg.Requests) {
					return
				}
				code, d, err := doRequest(ctx, client, cfg)
				if err != nil && ctx.Err() != nil {
					return
				}
				mu.Lock()
				latencies = append(latencies, float64(d.Microseconds())/1000)
				if err != nil {
					failed++
					errs[shorten(err.Error())]++
				} else {
					codes[code]++
					if code >= 200 && code < 400 {
						success++
					} else {
						failed++
					}
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	st := Stats{
		URL:           cfg.URL,
		Concurrency:   cfg.Concurrency,
		TotalRequests: int64(len(latencies)),
		Success:       success,
		Failed:        failed,
		ElapsedMs:     elapsed.Milliseconds(),
		StatusCodes:   codes,
	}
	if len(errs) > 0 {
		st.Errors = errs
	}
	if len(latencies) == 0 {
		return st
	}
	sort.Float64s(latencies)
	var sum float64
	for _, l := range latencies {
		sum += l
	}
	st.RPS = round(float64(len(latencies)) / elapsed.Seconds())
	st.MinMs = round(latencies[0])
	st.MeanMs = round(sum / float64(len(latencies)))
	st.P50Ms = round(percentile(latencies, 50))
	st.P90Ms = round(percentile(latencies, 90))
	st.P95Ms = round(percentile(latencies, 95))
	st.P99Ms = round(percentile(latencies, 99))
	st.MaxMs = round(latencies[len(latencies)-1])
	return st
}

func doRequest(ctx context.Context, client *http.Client, cfg Config) (int, time.Duration, error) {
	var body io.Reader
	if cfg.Body != "" {
		body = bytes.NewReader([]byte(cfg.Body))
	}
	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, body)
	if err != nil {
		return 0, 0, err
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, time.Since(start), err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode, time.Since(start), nil
}

func percentile(sorted []float64, p float64) float64 {
	idx := int(float64(len(sorted))*p/100+0.5) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func round(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }

func shorten(s string) string {
	if len(s) > 120 {
		return s[:120]
	}
	return s
}
