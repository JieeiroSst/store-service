package uploadservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/ticket-service/config"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const ReceiverPrefix = "ticket-user:"

func Receiver(userID int64) string { return ReceiverPrefix + strconv.FormatInt(userID, 10) }

type Client struct {
	base string
	key  string
	http *http.Client
}

var _ outbound.DocumentStore = (*Client)(nil)

func NewClient(cfg config.Config) *Client {
	return &Client{base: cfg.UploadServiceURL, key: cfg.UploadServiceKey,
		http: &http.Client{Timeout: cfg.PaymentServiceTimeout * 2}}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Service-Key", c.key)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: upload-service: %v", domain.ErrUpstreamUnavailable, err)
	}
	return resp, nil
}

func fail(resp *http.Response) error {
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	var e struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(raw, &e)
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return domain.ErrNotFound
	case resp.StatusCode >= 500 || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("%w: upload-service returned %d %s", domain.ErrUpstreamUnavailable, resp.StatusCode, e.Error)
	}
	return fmt.Errorf("%w: upload-service refused the file: %s", domain.ErrInvalid, e.Error)
}

func (c *Client) Put(ctx context.Context, userID int64, fileName string, pdf []byte) (string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(pdf); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	q := url.Values{"receiver_id": {Receiver(userID)}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/v1/upload?"+q.Encode(), &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fail(resp)
	}
	defer resp.Body.Close()
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil || out.ID == "" {
		return "", fmt.Errorf("%w: upload-service answered without a file id", domain.ErrUpstreamUnavailable)
	}
	return out.ID, nil
}

func (c *Client) Open(ctx context.Context, fileID string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/api/v1/upload/"+url.PathEscape(fileID)+"/content", nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fail(resp)
	}
	return resp.Body, resp.ContentLength, nil
}

func (c *Client) Delete(ctx context.Context, fileID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.base+"/api/v1/upload/"+url.PathEscape(fileID), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fail(resp)
	}
	resp.Body.Close()
	return nil
}
