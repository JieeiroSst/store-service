package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/draw-image-service/config"
	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"github.com/gin-gonic/gin"
)

const testID = "3f1c8a4e-2b7d-4c1e-9a5f-0d6b2e8c7a91"

type fakeCollage struct {
	got domain.CollageRequest
	err error
}

func (f *fakeCollage) Create(_ context.Context, req domain.CollageRequest) (*domain.Collage, error) {
	f.got = req
	for _, s := range req.Sources {
		io.Copy(io.Discard, s.Content)
	}
	if f.err != nil {
		return nil, f.err
	}
	return &domain.Collage{ID: testID, ContentType: "image/jpeg", Size: 4, ImageCount: len(req.Sources)}, nil
}

func (f *fakeCollage) Get(_ context.Context, id string) (*domain.Collage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &domain.Collage{ID: id, ContentType: "image/jpeg", Size: 4}, nil
}

func (f *fakeCollage) Open(ctx context.Context, id string) (*domain.Collage, io.ReadCloser, error) {
	c, err := f.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return c, io.NopCloser(bytes.NewReader([]byte("jpeg"))), nil
}

func newTestRouter(svc *fakeCollage, maxBytes int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(NewHandler(svc, &config.Config{Server: config.ServerConfig{MaxUploadBytes: maxBytes}}))
}

func multipartBody(t *testing.T, n int) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for i := 0; i < n; i++ {
		fw, err := mw.CreateFormFile("files", "../../etc/img.png")
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(fw, image.NewNRGBA(image.Rect(0, 0, 2, 2)))
	}
	mw.Close()
	return &buf, mw.FormDataContentType()
}

func upload(t *testing.T, svc *fakeCollage, maxBytes int64, target string, n int) *httptest.ResponseRecorder {
	body, ct := multipartBody(t, n)
	req := httptest.NewRequest(http.MethodPost, target, body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	newTestRouter(svc, maxBytes).ServeHTTP(rec, req)
	return rec
}

func get(svc *fakeCollage, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	newTestRouter(svc, 1<<20).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestUploadReturnsCreatedRecord(t *testing.T) {
	svc := &fakeCollage{}
	rec := upload(t, svc, 1<<20, "/upload?layout=row", 2)

	if rec.Code != http.StatusCreated {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	var body collageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ID != testID || body.URL != "/images/"+testID || rec.Header().Get("Location") != body.URL {
		t.Fatalf("body=%+v location=%q", body, rec.Header().Get("Location"))
	}
	if len(svc.got.Sources) != 2 || svc.got.Layout != domain.LayoutRow {
		t.Fatalf("request=%+v", svc.got)
	}
	if name := svc.got.Sources[0].Name; name != "img.png" {
		t.Fatalf("filename not sanitized: %q", name)
	}
}

func TestUploadPutsCenterFileFirst(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for _, f := range []struct{ field, name string }{
		{"files", "sale.png"}, {"center", "chair.png"}, {"files", "logo.png"},
	} {
		fw, _ := mw.CreateFormFile(f.field, f.name)
		png.Encode(fw, image.NewNRGBA(image.Rect(0, 0, 2, 2)))
	}
	mw.Close()

	svc := &fakeCollage{}
	req := httptest.NewRequest(http.MethodPost, "/upload?layout=center", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	newTestRouter(svc, 1<<20).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated || svc.got.Layout != domain.LayoutCenter {
		t.Fatalf("code=%d layout=%s", rec.Code, svc.got.Layout)
	}
	var names []string
	for _, s := range svc.got.Sources {
		names = append(names, s.Name)
	}
	if got := strings.Join(names, ","); got != "chair.png,sale.png,logo.png" {
		t.Fatalf("order=%s", got)
	}
}

func TestUploadValidation(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		maxBytes int64
		svcErr   error
		want     int
	}{
		{"bad layout", "/upload?layout=zigzag", 1 << 20, nil, http.StatusBadRequest},
		{"bad columns", "/upload?columns=0", 1 << 20, nil, http.StatusBadRequest},
		{"body too large", "/upload", 10, nil, http.StatusRequestEntityTooLarge},
		{"invalid image", "/upload", 1 << 20, domain.ErrInvalidImage, http.StatusBadRequest},
		{"too many images", "/upload", 1 << 20, domain.ErrTooManyImages, http.StatusRequestEntityTooLarge},
		{"storage down", "/upload", 1 << 20, errors.New("dial tcp minio-svc"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := upload(t, &fakeCollage{err: tt.svcErr}, tt.maxBytes, tt.target, 1)
			if rec.Code != tt.want {
				t.Fatalf("code=%d, want %d (%s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestInternalErrorsAreNotLeaked(t *testing.T) {
	rec := upload(t, &fakeCollage{err: errors.New("dial tcp minio-svc")}, 1<<20, "/upload", 1)
	if bytes.Contains(rec.Body.Bytes(), []byte("minio-svc")) {
		t.Fatalf("internal error leaked: %s", rec.Body.String())
	}
}

func TestGetImageStreamsJPEG(t *testing.T) {
	rec := get(&fakeCollage{}, "/images/"+testID)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/jpeg" || rec.Body.String() != "jpeg" {
		t.Fatalf("code=%d ct=%q body=%q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
}

func TestGetInfoReturnsRecord(t *testing.T) {
	rec := get(&fakeCollage{}, "/images/"+testID+"/info")
	var body collageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || rec.Code != http.StatusOK || body.ID != testID {
		t.Fatalf("code=%d body=%s err=%v", rec.Code, rec.Body.String(), err)
	}
}

func TestGetErrors(t *testing.T) {
	for _, tt := range []struct {
		err  error
		want int
	}{
		{domain.ErrInvalidID, http.StatusBadRequest},
		{domain.ErrNotFound, http.StatusNotFound},
	} {
		for _, target := range []string{"/images/x", "/images/x/info"} {
			if rec := get(&fakeCollage{err: tt.err}, target); rec.Code != tt.want {
				t.Fatalf("%s with %v: code=%d, want %d", target, tt.err, rec.Code, tt.want)
			}
		}
	}
}
