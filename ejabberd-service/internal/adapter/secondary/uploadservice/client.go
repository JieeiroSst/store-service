package uploadservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

const (
	basePath      = "/api/v1/upload"
	serviceHeader = "X-Service-Key"
)

type Client struct {
	baseURL    string
	serviceKey string
	http       *http.Client
}

var _ port.FileStorage = (*Client)(nil)

func New(cfg config.UploadServiceConfig) *Client {
	return &Client{
		baseURL:    strings.TrimRight(cfg.URL, "/"),
		serviceKey: cfg.ServiceKey,
		http:       &http.Client{Timeout: cfg.Timeout},
	}
}

type fileMeta struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(serviceHeader, c.serviceKey)
	return req, nil
}

func (c *Client) find(ctx context.Context, key string) (fileMeta, bool, error) {
	req, err := c.newRequest(ctx, http.MethodGet, basePath+"?limit=1&receiver_id="+url.QueryEscape(key), nil)
	if err != nil {
		return fileMeta{}, false, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fileMeta{}, false, fmt.Errorf("%w: %v", domain.ErrStorageUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fileMeta{}, false, statusError("list", resp)
	}
	var out struct {
		Files []fileMeta `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fileMeta{}, false, fmt.Errorf("%w: %v", domain.ErrStorageUnavailable, err)
	}
	if len(out.Files) == 0 {
		return fileMeta{}, false, nil
	}
	return out.Files[0], true, nil
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	_, ok, err := c.find(ctx, key)
	return ok, err
}

func (c *Client) Store(ctx context.Context, key, fileName string, size int64, body io.Reader) error {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		part, err := mw.CreateFormFile("file", fileName)
		if err == nil {
			_, err = io.Copy(part, body)
		}
		if err == nil {
			err = mw.Close()
		}
		pw.CloseWithError(err)
	}()
	req, err := c.newRequest(ctx, http.MethodPost, basePath+"?receiver_id="+url.QueryEscape(key), pr)
	if err != nil {
		pr.CloseWithError(err)
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		pr.CloseWithError(err)
		return fmt.Errorf("%w: %v", domain.ErrStorageUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return statusError("upload", resp)
	}
	return nil
}

func (c *Client) Fetch(ctx context.Context, key string) (domain.StoredFile, error) {
	meta, ok, err := c.find(ctx, key)
	if err != nil {
		return domain.StoredFile{}, err
	}
	if !ok {
		return domain.StoredFile{}, domain.ErrFileNotFound
	}
	req, err := c.newRequest(ctx, http.MethodGet, basePath+"/"+url.PathEscape(meta.ID)+"/content", nil)
	if err != nil {
		return domain.StoredFile{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return domain.StoredFile{}, fmt.Errorf("%w: %v", domain.ErrStorageUnavailable, err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return domain.StoredFile{}, statusError("download", resp)
	}
	out := domain.StoredFile{Name: meta.FileName, ContentType: meta.ContentType, Size: meta.Size, Body: resp.Body}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		out.ContentType = ct
	}
	if n, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64); err == nil {
		out.Size = n
	}
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil && params["filename"] != "" {
		out.Name = params["filename"]
	}
	return out, nil
}

func statusError(op string, resp *http.Response) error {
	var body struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&body)
	switch resp.StatusCode {
	case http.StatusRequestEntityTooLarge:
		return fmt.Errorf("%w: %s", domain.ErrFileTooLarge, body.Error)
	case http.StatusUnsupportedMediaType:
		return fmt.Errorf("%w: %s", domain.ErrUnsupportedFile, body.Error)
	case http.StatusNotFound:
		return domain.ErrFileNotFound
	case http.StatusBadRequest:
		return fmt.Errorf("%w: upload_service %s rejected the request: %s", domain.ErrInvalidSlot, op, body.Error)
	}
	return fmt.Errorf("%w: upload_service %s returned %d %s", domain.ErrStorageUnavailable, op, resp.StatusCode, body.Error)
}
