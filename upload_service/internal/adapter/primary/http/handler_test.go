package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

var pdf = []byte("%PDF-1.4\nbody")

// fakeUsecase serves a single file.
type fakeUsecase struct {
	file model.File
	body []byte
	err  error
}

func (f *fakeUsecase) Create(_ context.Context, in port.Upload) (*model.File, error) {
	if f.err != nil {
		return nil, f.err
	}
	b, _ := io.ReadAll(in.Body)
	f.body = b
	x := f.file
	x.ReceiverID, x.FileName = in.ReceiverID, in.FileName
	return &x, nil
}
func (f *fakeUsecase) Replace(ctx context.Context, _ string, in port.Upload) (*model.File, error) {
	return f.Create(ctx, in)
}
func (f *fakeUsecase) Get(context.Context, string) (*model.File, error) {
	if f.err != nil {
		return nil, f.err
	}
	x := f.file
	return &x, nil
}
func (f *fakeUsecase) Open(context.Context, string) (*port.Download, error) {
	if f.err != nil {
		return nil, f.err
	}
	x := f.file
	return &port.Download{File: &x, Body: io.NopCloser(bytes.NewReader(pdf))}, nil
}
func (f *fakeUsecase) List(context.Context, string, int, int) ([]model.File, int64, error) {
	return []model.File{f.file}, 1, f.err
}
func (f *fakeUsecase) Delete(context.Context, string) error { return f.err }

type stubValidator struct{ calls int }

func (s *stubValidator) Validate(_ context.Context, token string) (string, error) {
	s.calls++
	if token != "good" {
		return "", model.ErrUnauthenticated
	}
	return "u1", nil
}

func engine(uc *fakeUsecase, mode string, v *stubValidator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{Auth: config.AuthConfig{Mode: mode, CacheTTL: 5e9}, Upload: config.UploadConfig{MaxBytes: 1 << 10}}
	auth := NewAuthenticator(cfg, v)
	h := NewHandler(uc, cfg)
	e := gin.New()
	up := e.Group("/api/v1/upload", auth.Middleware())
	up.POST("", h.Create)
	up.GET("", h.List)
	up.GET("/:id", h.Get)
	up.GET("/:id/content", h.Content)
	up.PUT("/:id", h.Replace)
	up.DELETE("/:id", h.Delete)
	return e
}

func newUC() *fakeUsecase {
	return &fakeUsecase{file: model.File{ID: "f1", FileName: `re"port.pdf`, ContentType: "application/pdf", Size: int64(len(pdf)), ObjectKey: "SECRET-KEY"}}
}

func form(t *testing.T, field, name string, data []byte) (*bytes.Buffer, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile(field, name)
	_, _ = fw.Write(data)
	_ = w.Close()
	return &b, w.FormDataContentType()
}

func call(e *gin.Engine, method, path string, body io.Reader, ctype, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestEveryRouteRequiresAToken(t *testing.T) {
	e := engine(newUC(), config.AuthToken, &stubValidator{})
	for _, r := range [][2]string{{"POST", ""}, {"GET", ""}, {"GET", "/f1"}, {"GET", "/f1/content"}, {"PUT", "/f1"}, {"DELETE", "/f1"}} {
		if rec := call(e, r[0], "/api/v1/upload"+r[1], nil, "", ""); rec.Code != 401 {
			t.Errorf("%s %s without a token: %d, want 401", r[0], r[1], rec.Code)
		}
		if rec := call(e, r[0], "/api/v1/upload"+r[1], nil, "", "bad"); rec.Code != 401 {
			t.Errorf("%s %s with a bad token: %d, want 401", r[0], r[1], rec.Code)
		}
	}
}

func TestUploadReturnsTheIDAndNeverTheObjectKey(t *testing.T) {
	uc := newUC()
	body, ct := form(t, "file", "scan.pdf", pdf)
	rec := call(engine(uc, config.AuthToken, &stubValidator{}), "POST", "/api/v1/upload?receiver_id=patient:1", body, ct, "good")
	if rec.Code != 201 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["id"] != "f1" || got["receiver_id"] != "patient:1" || got["url"] != "/api/v1/upload/f1/content" {
		t.Errorf("response = %v", got)
	}
	if strings.Contains(rec.Body.String(), "SECRET-KEY") || got["object_key"] != nil {
		t.Errorf("internal object key leaked: %s", rec.Body)
	}
	if !bytes.Equal(uc.body, pdf) {
		t.Error("body not forwarded")
	}
}

func TestLegacyImageFieldStillWorksAndMissingFileIsA400(t *testing.T) {
	e := engine(newUC(), config.AuthOff, &stubValidator{})
	body, ct := form(t, "image", "a.pdf", pdf)
	if rec := call(e, "POST", "/api/v1/upload?receiver_id=r", body, ct, ""); rec.Code != 201 {
		t.Errorf("legacy field: %d", rec.Code)
	}
	body, ct = form(t, "other", "a.pdf", pdf)
	if rec := call(e, "POST", "/api/v1/upload?receiver_id=r", body, ct, ""); rec.Code != 400 {
		t.Errorf("wrong field: %d", rec.Code)
	}
}

func TestDownloadIsAnAttachmentWithSafeHeaders(t *testing.T) {
	rec := call(engine(newUC(), config.AuthToken, &stubValidator{}), "GET", "/api/v1/upload/f1/content", nil, "", "good")
	h := rec.Header()
	if rec.Code != 200 || !bytes.Equal(rec.Body.Bytes(), pdf) {
		t.Fatalf("%d %q", rec.Code, rec.Body)
	}
	if h.Get("Content-Type") != "application/pdf" || h.Get("X-Content-Type-Options") != "nosniff" ||
		!strings.HasPrefix(h.Get("Content-Disposition"), "attachment") || h.Get("Cache-Control") != "private, no-store" {
		t.Errorf("headers = %v", h)
	}
	// A quote in the file name must not be able to break out of the header.
	if strings.Contains(h.Get("Content-Disposition"), `re"port`) {
		t.Errorf("unescaped file name in %q", h.Get("Content-Disposition"))
	}
}

func TestErrorStatuses(t *testing.T) {
	cases := map[string]struct {
		err  error
		code int
	}{
		"not found":   {model.ErrNotFound, 404},
		"invalid":     {model.Invalid("x"), 400},
		"too large":   {model.ErrTooLarge, 413},
		"bad type":    {model.ErrUnsupportedType, 415},
		"storage out": {model.ErrUpstream, 503},
		"unexpected":  {io.ErrUnexpectedEOF, 500},
	}
	for name, tc := range cases {
		uc := newUC()
		uc.err = tc.err
		rec := call(engine(uc, config.AuthOff, &stubValidator{}), "GET", "/api/v1/upload/f1", nil, "", "")
		if rec.Code != tc.code {
			t.Errorf("%s: %d, want %d", name, rec.Code, tc.code)
		}
		if tc.code >= 500 && strings.Contains(rec.Body.String(), "unexpected EOF") {
			t.Errorf("%s: internal error text leaked: %s", name, rec.Body)
		}
	}
}

func TestTokenIsCachedBriefly(t *testing.T) {
	v := &stubValidator{}
	e := engine(newUC(), config.AuthToken, v)
	for i := 0; i < 3; i++ {
		call(e, "GET", "/api/v1/upload/f1", nil, "", "good")
	}
	if v.calls != 1 {
		t.Errorf("validator called %d times, want 1", v.calls)
	}
}
