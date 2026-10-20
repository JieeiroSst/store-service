package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const maxErrorBody = 4 << 10

type Client struct {
	name    string
	baseURL string
	http    *http.Client
	opts    Options
}

type Options struct {
	DropUserToken     bool
	BasicAuthUser     string
	BasicAuthPassword string
	Headers           map[string]string
}

type Request struct {
	Method string
	Path   string
	Query  url.Values
	Body   any
	Header http.Header
}

type File struct {
	Field    string
	Upload   graphql.Upload
	Filename string
}

func New(name, baseURL string, timeout time.Duration, opts ...Options) *Client {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	return &Client{
		opts:    o,
		name:    name,
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *Client) Name() string { return c.name }

func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	return c.Do(ctx, Request{Method: http.MethodGet, Path: path, Query: query}, out)
}

func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	return c.Do(ctx, Request{Method: http.MethodPost, Path: path, Body: body}, out)
}

func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	return c.Do(ctx, Request{Method: http.MethodPut, Path: path, Body: body}, out)
}

func (c *Client) Patch(ctx context.Context, path string, body, out any) error {
	return c.Do(ctx, Request{Method: http.MethodPatch, Path: path, Body: body}, out)
}

func (c *Client) Delete(ctx context.Context, path string, query url.Values, out any) error {
	return c.Do(ctx, Request{Method: http.MethodDelete, Path: path, Query: query}, out)
}

func (c *Client) Do(ctx context.Context, req Request, out any) error {
	var body io.Reader
	contentType := ""
	if req.Body != nil {
		payload, err := encodeBody(req.Body)
		if err != nil {
			return c.fail(ctx, auth.CodeBadRequest, http.StatusBadRequest, "invalid request body")
		}
		body = bytes.NewReader(payload)
		contentType = "application/json"
	}
	return c.send(ctx, req, body, contentType, out)
}

func (c *Client) Raw(ctx context.Context, req Request, upload graphql.Upload, out any) error {
	contentType := upload.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return c.send(ctx, req, upload.File, contentType, out)
}

func (c *Client) Multipart(ctx context.Context, req Request, fields map[string][]string, files []File, out any) error {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	for k, vs := range fields {
		for _, v := range vs {
			if err := w.WriteField(k, v); err != nil {
				return c.fail(ctx, auth.CodeBadRequest, http.StatusBadRequest, "invalid multipart field")
			}
		}
	}
	for _, f := range files {
		name := f.Filename
		if name == "" {
			name = f.Upload.Filename
		}
		part, err := w.CreateFormFile(f.Field, name)
		if err != nil {
			return c.fail(ctx, auth.CodeBadRequest, http.StatusBadRequest, "invalid file")
		}
		if _, err := io.Copy(part, f.Upload.File); err != nil {
			return c.fail(ctx, auth.CodeBadRequest, http.StatusBadRequest, "cannot read file")
		}
	}
	if err := w.Close(); err != nil {
		return c.fail(ctx, auth.CodeBadRequest, http.StatusBadRequest, "invalid multipart body")
	}
	return c.send(ctx, req, buf, w.FormDataContentType(), out)
}

func (c *Client) send(ctx context.Context, req Request, body io.Reader, contentType string, out any) error {
	if c.baseURL == "" {
		return c.fail(ctx, auth.CodeServiceUnavailable, 0, "service is not configured")
	}
	target := c.baseURL + req.Path
	if len(req.Query) > 0 {
		target += "?" + req.Query.Encode()
	}
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return c.fail(ctx, auth.CodeBadRequest, http.StatusBadRequest, "invalid request")
	}
	for k, vs := range req.Header {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	httpReq.Header.Set("Accept", "application/json")
	client := auth.ClientFromContext(ctx)
	if client.IP != "" {
		httpReq.Header.Set("X-Forwarded-For", client.IP)
		httpReq.Header.Set("X-Real-IP", client.IP)
		if client.UserAgent != "" {
			httpReq.Header.Set("User-Agent", client.UserAgent)
		}
	}
	if client.DeviceID != "" {
		httpReq.Header.Set("X-Device-Id", client.DeviceID)
	}
	if user, ok := auth.UserFromContext(ctx); ok {
		if !c.opts.DropUserToken {
			httpReq.Header.Set("Authorization", "Bearer "+user.Token)
		}
		httpReq.Header.Set("X-User-Id", user.ID)
	}
	for k, v := range c.opts.Headers {
		if v != "" {
			httpReq.Header.Set(k, v)
		}
	}
	if c.opts.BasicAuthUser != "" {
		httpReq.SetBasicAuth(c.opts.BasicAuthUser, c.opts.BasicAuthPassword)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		logger.Error(ctx, "%s %s %s: %v", c.name, method, req.Path, err)
		if errors.Is(err, context.DeadlineExceeded) {
			return c.fail(ctx, auth.CodeServiceUnavailable, http.StatusGatewayTimeout, "service timed out")
		}
		return c.fail(ctx, auth.CodeServiceUnavailable, http.StatusBadGateway, "service is unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		if resp.StatusCode >= 500 {
			logger.Error(ctx, "%s %s %s: status %d: %s", c.name, method, req.Path, resp.StatusCode, string(raw))
		}
		return c.fail(ctx, codeForStatus(resp.StatusCode), resp.StatusCode, upstreamMessage(resp.StatusCode, raw))
	}

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return redirectBody(resp.Header.Get("Location"), out)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if a, ok := out.(*any); ok && !strings.Contains(resp.Header.Get("Content-Type"), "json") {
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.fail(ctx, auth.CodeServiceUnavailable, http.StatusBadGateway, "cannot read service response")
		}
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
			dec := json.NewDecoder(bytes.NewReader(trimmed))
			dec.UseNumber()
			if dec.Decode(a) == nil {
				return nil
			}
		}
		*a = string(raw)
		return nil
	}
	if s, ok := out.(*string); ok {
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.fail(ctx, auth.CodeServiceUnavailable, http.StatusBadGateway, "cannot read service response")
		}
		*s = string(raw)
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(out); err != nil && !errors.Is(err, io.EOF) {
		logger.Error(ctx, "%s %s %s: decode: %v", c.name, method, req.Path, err)
		return c.fail(ctx, auth.CodeServiceUnavailable, http.StatusBadGateway, "unexpected service response")
	}
	return nil
}

func (c *Client) fail(ctx context.Context, code string, status int, message string) error {
	ext := map[string]interface{}{"code": code, "service": c.name}
	if status > 0 {
		ext["status"] = status
	}
	return &gqlerror.Error{Path: graphql.GetPath(ctx), Message: message, Extensions: ext}
}

func codeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return auth.CodeBadRequest
	case http.StatusUnauthorized:
		return auth.CodeUnauthenticated
	case http.StatusForbidden:
		return auth.CodeForbidden
	case http.StatusNotFound:
		return auth.CodeNotFound
	case http.StatusConflict:
		return auth.CodeConflict
	case http.StatusTooManyRequests:
		return auth.CodeTooManyRequests
	}
	if status >= 500 {
		return auth.CodeServiceUnavailable
	}
	return auth.CodeBadRequest
}

func upstreamMessage(status int, raw []byte) string {
	if status >= 500 {
		return "service error"
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err == nil {
		for _, key := range []string{"message", "error", "msg", "detail", "error_description"} {
			if s, ok := body[key].(string); ok && s != "" {
				return s
			}
		}
		if e, ok := body["error"].(map[string]any); ok {
			if s, ok := e["message"].(string); ok && s != "" {
				return s
			}
		}
	}
	if s := strings.TrimSpace(string(raw)); s != "" && len(s) < 300 && !strings.HasPrefix(s, "<") {
		return s
	}
	return http.StatusText(status)
}

type RawSegment string

func Raw(path string) RawSegment {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			continue
		}
		out = append(out, url.PathEscape(p))
	}
	return RawSegment(strings.Join(out, "/"))
}

func P(format string, args ...any) string {
	escaped := make([]any, len(args))
	for i, a := range args {
		if r, ok := a.(RawSegment); ok {
			escaped[i] = string(r)
			continue
		}
		escaped[i] = url.PathEscape(fmt.Sprint(a))
	}
	return fmt.Sprintf(format, escaped...)
}

func encodeBody(body any) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return json.Marshal(stripNulls(v))
}

func stripNulls(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if val == nil {
				delete(t, k)
				continue
			}
			t[k] = stripNulls(val)
		}
	case []any:
		for i, val := range t {
			t[i] = stripNulls(val)
		}
	}
	return v
}

func redirectBody(location string, out any) error {
	if out == nil {
		return nil
	}
	payload, _ := json.Marshal(map[string]string{"location": location, "target_url": location})
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	return dec.Decode(out)
}
