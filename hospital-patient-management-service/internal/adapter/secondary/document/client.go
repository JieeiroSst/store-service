package document

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
	"strings"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

const receiverPrefix = "patient:"

// Client talks to upload-service. Files move through this service in memory
// only; nothing is written to the hospital's disk or database.
type Client struct {
	base string
	http *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		base: cfg.Upload.BaseURL,
		// File transfers take longer than the JSON calls the timeout is tuned for.
		http: &http.Client{Timeout: cfg.Upload.Timeout * 6},
	}
}

func (c *Client) Enabled() bool { return c.base != "" }

func receiver(patientID int32) string { return receiverPrefix + strconv.Itoa(int(patientID)) }

type fileDTO struct {
	ID          string    `json:"id"`
	ReceiverID  string    `json:"receiver_id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

func (f fileDTO) toModel(patientID int32) model.Document {
	return model.Document{ID: f.ID, PatientID: patientID, FileName: f.FileName, ContentType: f.ContentType, Size: f.Size, CreatedAt: f.CreatedAt}
}

func (c *Client) Upload(ctx context.Context, patientID int32, in port.DocumentUpload) (*model.Document, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("file", in.FileName)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, in.Body); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	var out fileDTO
	q := url.Values{"receiver_id": {receiver(patientID)}}
	if err := c.do(ctx, http.MethodPost, "/api/v1/upload?"+q.Encode(), &body, w.FormDataContentType(), &out); err != nil {
		return nil, err
	}
	d := out.toModel(patientID)
	return &d, nil
}

func (c *Client) List(ctx context.Context, patientID int32, limit, offset int) ([]model.Document, int64, error) {
	q := url.Values{"receiver_id": {receiver(patientID)}, "limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}}
	var out struct {
		Files      []fileDTO `json:"files"`
		TotalCount int64     `json:"total_count"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/v1/upload?"+q.Encode(), nil, "", &out); err != nil {
		return nil, 0, err
	}
	docs := make([]model.Document, len(out.Files))
	for i, f := range out.Files {
		docs[i] = f.toModel(patientID)
	}
	return docs, out.TotalCount, nil
}

// meta fetches a file's record and proves it belongs to this patient. Without
// this check, any known file id could be read or deleted through any patient.
func (c *Client) meta(ctx context.Context, patientID int32, docID string) (*model.Document, error) {
	var f fileDTO
	if err := c.do(ctx, http.MethodGet, "/api/v1/upload/"+url.PathEscape(docID), nil, "", &f); err != nil {
		return nil, err
	}
	if f.ReceiverID != receiver(patientID) {
		return nil, model.ErrNotFound
	}
	d := f.toModel(patientID)
	return &d, nil
}

func (c *Client) Open(ctx context.Context, patientID int32, docID string) (*port.DocumentDownload, error) {
	doc, err := c.meta(ctx, patientID, docID)
	if err != nil {
		return nil, err
	}
	resp, err := c.send(ctx, http.MethodGet, "/api/v1/upload/"+url.PathEscape(docID)+"/content", nil, "")
	if err != nil {
		return nil, err
	}
	if err := statusErr(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return &port.DocumentDownload{Document: doc, Body: resp.Body}, nil
}

func (c *Client) Delete(ctx context.Context, patientID int32, docID string) error {
	if _, err := c.meta(ctx, patientID, docID); err != nil {
		return err
	}
	return c.do(ctx, http.MethodDelete, "/api/v1/upload/"+url.PathEscape(docID), nil, "", nil)
}

func (c *Client) send(ctx context.Context, method, path string, body io.Reader, ctype string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return nil, err
	}
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	// Act as the caller: upload-service authenticates the same user token.
	if t := port.BearerFrom(ctx); t != "" {
		req.Header.Set("Authorization", "Bearer "+t)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: upload-service: %v", model.ErrUpstream, err)
	}
	return resp, nil
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, ctype string, out any) error {
	resp, err := c.send(ctx, method, path, body, ctype)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := statusErr(resp); err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}

// statusErr sorts upload-service answers: input it refused is the caller's
// problem, while a refusal of our own credentials is a deployment problem.
func statusErr(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	var p struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&p)
	msg := strings.TrimSpace(p.Error)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return model.ErrNotFound
	case http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType, http.StatusBadRequest:
		return model.Invalid("%s", msg)
	case http.StatusUnauthorized, http.StatusForbidden:
		logrus.Errorf("upload-service rejected our credentials (status %d): is the caller's token valid there?", resp.StatusCode)
		return fmt.Errorf("%w: upload-service refused the request", model.ErrUpstream)
	default:
		return fmt.Errorf("%w: upload-service status %d: %s", model.ErrUpstream, resp.StatusCode, msg)
	}
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.DocumentGateway)))),
)
