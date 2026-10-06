package serpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

const maxBodyBytes = 32 << 20

type Client struct {
	baseURL    string
	apiKey     string
	maxRetries int
	http       *http.Client
	backoff    time.Duration
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:    cfg.SerpAPI.BaseURL,
		apiKey:     cfg.SerpAPI.APIKey,
		maxRetries: cfg.SerpAPI.MaxRetries,
		http:       &http.Client{Timeout: cfg.SerpAPI.Timeout},
		backoff:    300 * time.Millisecond,
	}
}

func (c *Client) Search(ctx context.Context, req domain.SearchRequest) (domain.SearchResult, error) {
	q := url.Values{}
	for k, v := range req.Params {
		q.Set(k, v)
	}
	q.Set("engine", req.Engine)
	q.Set("output", string(req.Output))
	if req.NoCache {
		q.Set("no_cache", "true")
	}
	if req.Async {
		q.Set("async", "true")
	}
	return c.result(ctx, "/search", q, req.Output)
}

func (c *Client) Archive(ctx context.Context, searchID string, output domain.Output) (domain.SearchResult, error) {
	return c.result(ctx, "/searches/"+url.PathEscape(searchID)+"."+string(output), url.Values{}, output)
}

func (c *Client) Account(ctx context.Context) (domain.Account, error) {
	var acc domain.Account
	body, _, err := c.get(ctx, "/account.json", url.Values{}, true)
	if err != nil {
		return acc, err
	}
	if err := json.Unmarshal(body, &acc); err != nil {
		return acc, &domain.UpstreamError{Msg: "decode account: " + err.Error()}
	}
	return acc, nil
}

func (c *Client) Locations(ctx context.Context, lq domain.LocationQuery) ([]domain.Location, error) {
	q := url.Values{}
	if lq.Q != "" {
		q.Set("q", lq.Q)
	}
	q.Set("limit", strconv.Itoa(lq.Limit))
	body, _, err := c.get(ctx, "/locations.json", q, false)
	if err != nil {
		return nil, err
	}
	var locs []domain.Location
	if err := json.Unmarshal(body, &locs); err != nil {
		return nil, &domain.UpstreamError{Msg: "decode locations: " + err.Error()}
	}
	return locs, nil
}

func (c *Client) UploadImage(ctx context.Context, img domain.ImageUpload) (domain.UploadedImage, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	err := mw.WriteField("api_key", c.apiKey)
	var part io.Writer
	if err == nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename=%q`, img.Filename))
		h.Set("Content-Type", img.ContentType)
		part, err = mw.CreatePart(h)
	}
	if err == nil {
		_, err = part.Write(img.Data)
	}
	if err == nil {
		err = mw.Close()
	}
	if err != nil {
		return domain.UploadedImage{}, &domain.UpstreamError{Msg: "build upload: " + err.Error()}
	}
	payload, ctype := buf.Bytes(), mw.FormDataContentType()

	body, _, err := c.send(ctx, func(ctx context.Context) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/image", bytes.NewReader(payload))
		if err == nil {
			req.Header.Set("Content-Type", ctype)
		}
		return req, err
	})
	if err != nil {
		return domain.UploadedImage{}, err
	}
	var out struct {
		ImageID string `json:"image_id"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return domain.UploadedImage{}, &domain.UpstreamError{Msg: "decode upload: " + err.Error()}
	}
	if out.ImageID == "" {
		if out.Error != "" {
			return domain.UploadedImage{}, domain.Invalid("%s", out.Error)
		}
		return domain.UploadedImage{}, &domain.UpstreamError{Msg: "upload returned no image_id"}
	}
	return domain.UploadedImage{ImageID: out.ImageID, Message: out.Message}, nil
}

func (c *Client) result(ctx context.Context, path string, q url.Values, output domain.Output) (domain.SearchResult, error) {
	body, ctype, err := c.get(ctx, path, q, true)
	if err != nil {
		return domain.SearchResult{}, err
	}
	res := domain.SearchResult{Body: body, ContentType: ctype}
	if output.IsJSON() {
		var meta struct {
			SearchMetadata struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"search_metadata"`
		}
		if err := json.Unmarshal(body, &meta); err != nil {
			return res, &domain.UpstreamError{Msg: "decode search result: " + err.Error()}
		}
		res.SearchID = meta.SearchMetadata.ID
		res.Status = meta.SearchMetadata.Status
	}
	return res, nil
}

func (c *Client) get(ctx context.Context, path string, q url.Values, withKey bool) ([]byte, string, error) {
	if withKey {
		q.Set("api_key", c.apiKey)
	}
	u := c.baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return c.send(ctx, func(ctx context.Context) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err == nil {
			req.Header.Set("Accept", "application/json, text/html, text/markdown")
		}
		return req, err
	})
}

func (c *Client) send(ctx context.Context, build func(context.Context) (*http.Request, error)) ([]byte, string, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(c.backoff << (attempt - 1)):
			}
		}
		body, ctype, err := c.do(ctx, build)
		if err == nil {
			return body, ctype, nil
		}
		lastErr = err
		if !retryable(err) || ctx.Err() != nil {
			break
		}
	}
	return nil, "", lastErr
}

func (c *Client) do(ctx context.Context, build func(context.Context) (*http.Request, error)) ([]byte, string, error) {
	req, err := build(ctx)
	if err != nil {
		return nil, "", &domain.UpstreamError{Msg: "build request: " + redact(err)}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		return nil, "", &domain.UpstreamError{Msg: redact(err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, "", &domain.UpstreamError{Status: resp.StatusCode, Msg: "read body: " + redact(err)}
	}
	if len(body) > maxBodyBytes {
		return nil, "", &domain.UpstreamError{Status: resp.StatusCode, Msg: "response too large"}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, resp.Header.Get("Content-Type"), nil
	}
	return nil, "", statusError(resp.StatusCode, errorMessage(body))
}

func statusError(status int, msg string) error {
	switch {
	case status == http.StatusBadRequest:
		return domain.Invalid("%s", msg)
	case status == http.StatusNotFound:
		return fmt.Errorf("%w: %s", domain.ErrNotFound, msg)
	case status == http.StatusGone:
		return fmt.Errorf("%w: %s", domain.ErrExpired, msg)
	case status == http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", domain.ErrRateLimited, msg)
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return &domain.UpstreamError{Status: status, Msg: "API key rejected: " + msg}
	default:
		return &domain.UpstreamError{Status: status, Msg: msg}
	}
}

func retryable(err error) bool {
	var ue *domain.UpstreamError
	if !errors.As(err, &ue) {
		return false
	}
	return ue.Status == 0 || ue.Status >= 500
}

func errorMessage(body []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error != "" {
		return e.Error
	}
	s := strings.TrimSpace(string(body))
	if len(s) > 200 {
		s = s[:200]
	}
	if s == "" {
		s = "empty response"
	}
	return s
}

func redact(err error) string {
	var ue *url.Error
	if errors.As(err, &ue) {
		return ue.Op + ": " + ue.Err.Error()
	}
	return err.Error()
}
