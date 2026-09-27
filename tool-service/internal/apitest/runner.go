package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TestCase struct {
	Name               string            `json:"name"`
	Method             string            `json:"method"`
	Path               string            `json:"path"`
	Headers            map[string]string `json:"headers,omitempty"`
	Body               any               `json:"body,omitempty"`
	ExpectStatus       int               `json:"expect_status"`
	ExpectBodyContains []string          `json:"expect_body_contains,omitempty"`

	ExpectJSON   map[string]any `json:"expect_json,omitempty"`
	MaxLatencyMs int64          `json:"max_latency_ms,omitempty"`

	SpecQuote string `json:"spec_quote,omitempty"`
}

type Result struct {
	Name       string   `json:"name"`
	Passed     bool     `json:"passed"`
	Status     int      `json:"status"`
	LatencyMs  int64    `json:"latency_ms"`
	Failures   []string `json:"failures,omitempty"`
	BodySample string   `json:"body_sample,omitempty"`
}

type Report struct {
	Total   int      `json:"total"`
	Passed  int      `json:"passed"`
	Failed  int      `json:"failed"`
	Results []Result `json:"results"`
}

const maxBody = 1 << 20

func Run(ctx context.Context, client *http.Client, baseURL string, cases []TestCase) Report {
	baseURL = strings.TrimRight(baseURL, "/")
	rep := Report{Total: len(cases)}
	for _, tc := range cases {
		res := runOne(ctx, client, baseURL, tc)
		if res.Passed {
			rep.Passed++
		} else {
			rep.Failed++
		}
		rep.Results = append(rep.Results, res)
	}
	return rep
}

func runOne(ctx context.Context, client *http.Client, baseURL string, tc TestCase) Result {
	res := Result{Name: tc.Name}
	method := strings.ToUpper(tc.Method)
	if method == "" {
		method = http.MethodGet
	}
	path := tc.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var reader io.Reader
	if tc.Body != nil {
		b, err := json.Marshal(tc.Body)
		if err != nil {
			res.Failures = append(res.Failures, "invalid body: "+err.Error())
			return res
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reader)
	if err != nil {
		res.Failures = append(res.Failures, "build request: "+err.Error())
		return res
	}
	if tc.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range tc.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	res.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		res.Failures = append(res.Failures, "request failed: "+err.Error())
		return res
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	res.Status = resp.StatusCode
	res.BodySample = truncate(string(body), 300)

	expect := tc.ExpectStatus
	if expect == 0 {
		expect = http.StatusOK
	}
	if resp.StatusCode != expect {
		res.Failures = append(res.Failures, fmt.Sprintf("status: want %d, got %d", expect, resp.StatusCode))
	}
	for _, s := range tc.ExpectBodyContains {
		if !strings.Contains(string(body), s) {
			res.Failures = append(res.Failures, fmt.Sprintf("body does not contain %q", s))
		}
	}
	if len(tc.ExpectJSON) > 0 {
		var parsed any
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.UseNumber()
		if err := dec.Decode(&parsed); err != nil {
			res.Failures = append(res.Failures, "response is not valid JSON")
		} else {
			for path, want := range tc.ExpectJSON {
				exp, err := ResolveExpected(want)
				if err != nil {
					res.Failures = append(res.Failures, fmt.Sprintf("json path %q: invalid expected value: %v", path, err))
					continue
				}
				got, ok := lookup(parsed, path)
				if !ok {
					res.Failures = append(res.Failures, fmt.Sprintf("json path %q not found", path))
				} else if !equalValues(got, exp) {
					res.Failures = append(res.Failures, fmt.Sprintf("json path %q: want %s, got %v", path, showValue(exp), got))
				}
			}
		}
	}
	if tc.MaxLatencyMs > 0 && res.LatencyMs > tc.MaxLatencyMs {
		res.Failures = append(res.Failures, fmt.Sprintf("latency %dms exceeds %dms", res.LatencyMs, tc.MaxLatencyMs))
	}
	res.Passed = len(res.Failures) == 0
	return res
}

func lookup(v any, path string) (any, bool) {
	for _, part := range strings.Split(path, ".") {
		switch t := v.(type) {
		case map[string]any:
			next, ok := t[part]
			if !ok {
				return nil, false
			}
			v = next
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(t) {
				return nil, false
			}
			v = t[i]
		default:
			return nil, false
		}
	}
	return v, true
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
