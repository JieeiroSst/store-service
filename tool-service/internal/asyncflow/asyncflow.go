package asyncflow

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
)

type Idempotency struct {
	Header string `json:"header"`
	Key    string `json:"key"`
	Repeat int    `json:"repeat"`
}

type Spec struct {
	Trigger apitest.TestCase `json:"trigger"`

	Poll        apitest.TestCase `json:"poll"`
	TimeoutSec  int              `json:"timeout_sec"`
	IntervalMs  int              `json:"interval_ms"`
	Idempotency *Idempotency     `json:"idempotency,omitempty"`
}

type Result struct {
	Passed        bool     `json:"passed"`
	TriggerStatus int      `json:"trigger_status"`
	Settled       bool     `json:"settled"`
	SettleMs      int64    `json:"settle_ms"`
	Polls         int      `json:"polls"`
	Replays       []int    `json:"replay_statuses,omitempty"`
	Failures      []string `json:"failures,omitempty"`
}

func Run(ctx context.Context, c *http.Client, baseURL string, s Spec) Result {
	base := strings.TrimRight(baseURL, "/")
	res := Result{}
	fail := func(f string, a ...any) { res.Failures = append(res.Failures, fmt.Sprintf(f, a...)) }

	timeout := time.Duration(clamp(s.TimeoutSec, 1, 300, 30)) * time.Second
	interval := time.Duration(clamp(s.IntervalMs, 50, 10000, 500)) * time.Millisecond

	hdr := map[string]string{}
	for k, v := range s.Trigger.Headers {
		hdr[k] = v
	}
	sends := 1
	if s.Idempotency != nil {
		sends = clamp(s.Idempotency.Repeat, 2, 5, 2)
		name := s.Idempotency.Header
		if name == "" {
			name = "Idempotency-Key"
		}
		key := s.Idempotency.Key
		if key == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			key = "qa-" + hex.EncodeToString(b)
		}
		hdr[name] = key
	}

	start := time.Now()
	var bodies []string
	for i := 0; i < sends; i++ {
		status, body, err := send(ctx, c, s.Trigger.Method, base+path(s.Trigger.Path), hdr, s.Trigger.Body)
		if err != nil {
			fail("trigger %d failed: %v", i+1, err)
			return res
		}
		if i == 0 {
			res.TriggerStatus = status
		} else {
			res.Replays = append(res.Replays, status)
		}
		bodies = append(bodies, body)
		if status < 200 || status >= 300 {
			if i == 0 {
				fail("trigger returned %d: %s", status, trunc(body))
				return res
			}
			if status != http.StatusConflict {
				fail("replay %d returned %d (want the original result or 409)", i+1, status)
			}
		}
	}
	if s.Idempotency != nil {
		for i := 1; i < len(bodies); i++ {
			if res.Replays[i-1] >= 200 && res.Replays[i-1] < 300 && bodies[i] != bodies[0] {
				fail("replay %d returned a different result than the first request (not idempotent)", i+1)
			}
		}
	}

	poll := s.Poll
	if poll.Path == "" {
		fail("poll.path is required")
		return res
	}
	deadline := start.Add(timeout)
	for {
		rep := apitest.Run(ctx, c, base, []apitest.TestCase{poll})
		res.Polls++
		if rep.Passed == 1 {
			res.Settled, res.SettleMs = true, time.Since(start).Milliseconds()
			break
		}
		if time.Now().Add(interval).After(deadline) || ctx.Err() != nil {
			last := rep.Results[0]
			fail("effect not visible after %s (%d polls): %s", timeout, res.Polls, strings.Join(last.Failures, "; "))
			break
		}
		select {
		case <-time.After(interval):
		case <-ctx.Done():
		}
	}
	res.Passed = len(res.Failures) == 0 && res.Settled
	return res
}

func send(ctx context.Context, c *http.Client, method, url string, hdr map[string]string, body any) (int, string, error) {
	if method == "" {
		method = http.MethodPost
	}
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, "", err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, r)
	if err != nil {
		return 0, "", err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, string(b), nil
}

func path(p string) string { return "/" + strings.TrimLeft(p, "/") }

func trunc(s string) string {
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

func clamp(v, lo, hi, def int) int {
	if v == 0 {
		return def
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
