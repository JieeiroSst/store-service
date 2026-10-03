package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ctxKey struct{}

var forwardedHeaders = []string{"Authorization", "Cookie", "X-Request-Id", "Accept-Language"}

func WithIncomingHeaders(ctx context.Context, h http.Header) context.Context {
	return context.WithValue(ctx, ctxKey{}, h.Clone())
}

func ForwardHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(WithIncomingHeaders(r.Context(), r.Header)))
	})
}

type Client struct {
	Service string
	BaseURL string
	HTTP    *http.Client
	Forward []string
}

func New(service, baseURL string, httpClient *http.Client, forward ...string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		Service: service,
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    httpClient,
		Forward: append(append([]string{}, forwardedHeaders...), forward...),
	}
}

type Error struct {
	Service string
	Path    string
	Status  int
	Body    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s GET %s: %d %s", e.Service, e.Path, e.Status, e.Body)
}

func (c *Client) Get(ctx context.Context, path string, query url.Values, headers http.Header, unwrap string, out any) error {
	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if h, ok := ctx.Value(ctxKey{}).(http.Header); ok {
		for _, k := range c.Forward {
			if v := h.Values(k); len(v) > 0 {
				req.Header[http.CanonicalHeaderKey(k)] = v
			}
		}
	}
	for k, v := range headers {
		req.Header[k] = v
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%s GET %s: %w", c.Service, path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("%s GET %s: %w", c.Service, path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &Error{Service: c.Service, Path: path, Status: resp.StatusCode, Body: errorText(body)}
	}
	return Decode(body, unwrap, out)
}

func Decode(body []byte, unwrap string, out any) error {
	raw := json.RawMessage(bytes.TrimSpace(body))
	if len(raw) == 0 {
		return nil
	}
	if unwrap != "" {
		for _, key := range strings.Split(unwrap, ".") {
			if bytes.Equal(raw, []byte("null")) {
				return nil
			}
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(raw, &obj); err != nil {
				return fmt.Errorf("unwrap %q: %w", unwrap, err)
			}
			next, ok := obj[key]
			if !ok {
				return fmt.Errorf("unwrap %q: key %q not in response", unwrap, key)
			}
			raw = next
		}
	}
	if err := json.Unmarshal(raw, out); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return fmt.Errorf("decode response field %q: got JSON %s, want %s", typeErr.Field, typeErr.Value, typeErr.Type)
		}
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func errorText(body []byte) string {
	var obj map[string]any
	if json.Unmarshal(body, &obj) == nil {
		for _, k := range []string{"error", "message", "msg", "detail"} {
			if v, ok := obj[k]; ok {
				return fmt.Sprint(v)
			}
		}
	}
	s := strings.TrimSpace(string(body))
	if len(s) > 300 {
		s = s[:300] + "…"
	}
	return s
}

// EscapePath escapes each segment of a multi-segment path parameter (a
// wildcard route), keeping the slashes between them.
func EscapePath(p string) string {
	parts := strings.Split(strings.Trim(p, "/"), "/")
	for i, s := range parts {
		parts[i] = url.PathEscape(s)
	}
	return strings.Join(parts, "/")
}
