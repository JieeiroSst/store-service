package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
)

var pdf = []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n")
var png = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89")

type memMeta struct {
	rows       map[string]model.File
	failInsert error
	failUpdate error
}

func (m *memMeta) Insert(_ context.Context, f *model.File) error {
	if m.failInsert != nil {
		return m.failInsert
	}
	m.rows[f.ID] = *f
	return nil
}
func (m *memMeta) Get(_ context.Context, id string) (*model.File, error) {
	f, ok := m.rows[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return &f, nil
}
func (m *memMeta) List(_ context.Context, rid string, limit, offset int) ([]model.File, int64, error) {
	var out []model.File
	for _, f := range m.rows {
		if f.ReceiverID == rid {
			out = append(out, f)
		}
	}
	return out, int64(len(out)), nil
}
func (m *memMeta) ReplaceContent(_ context.Context, f *model.File) error {
	if m.failUpdate != nil {
		return m.failUpdate
	}
	m.rows[f.ID] = *f
	return nil
}
func (m *memMeta) Delete(_ context.Context, id string) error {
	if _, ok := m.rows[id]; !ok {
		return model.ErrNotFound
	}
	delete(m.rows, id)
	return nil
}

type memObjects struct {
	data       map[string][]byte
	failDelete error
}

func (m *memObjects) Put(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	b, _ := io.ReadAll(body)
	m.data[key] = b
	return nil
}
func (m *memObjects) Get(_ context.Context, key string) (io.ReadCloser, error) {
	b, ok := m.data[key]
	if !ok {
		return nil, model.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func (m *memObjects) Delete(_ context.Context, key string) error {
	if m.failDelete != nil {
		return m.failDelete
	}
	delete(m.data, key)
	return nil
}

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { c.t = c.t.Add(time.Second); return c.t }

func newSvc(maxBytes int64) (port.FileUsecase, *memMeta, *memObjects) {
	meta := &memMeta{rows: map[string]model.File{}}
	obj := &memObjects{data: map[string][]byte{}}
	cfg := &config.Config{Upload: config.UploadConfig{MaxBytes: maxBytes, AllowedTypes: []string{"application/pdf", "image/png"}}}
	return NewFileService(meta, obj, &fixedClock{t: time.Unix(1_700_000_000, 0)}, cfg), meta, obj
}

func up(name string, data []byte) port.Upload {
	return port.Upload{ReceiverID: "patient:1", FileName: name, Body: bytes.NewReader(data)}
}

func TestCreateStoresBytesAndMetadata(t *testing.T) {
	svc, meta, obj := newSvc(1 << 20)
	f, err := svc.Create(context.Background(), up("scan.pdf", pdf))
	if err != nil {
		t.Fatal(err)
	}
	if f.ContentType != "application/pdf" || f.Size != int64(len(pdf)) || len(f.SHA256) != 64 || f.ID == "" || f.ObjectKey == "" {
		t.Errorf("file = %+v", f)
	}
	if !bytes.Equal(obj.data[f.ObjectKey], pdf) || meta.rows[f.ID].ReceiverID != "patient:1" {
		t.Error("bytes or metadata not stored")
	}
	if f.ObjectKey == f.ID {
		t.Error("the object key must not be guessable from the public id")
	}
}

func TestTypeComesFromTheBytesNotTheName(t *testing.T) {
	svc, _, _ := newSvc(1 << 20)
	// A script renamed to .pdf is still a script.
	_, err := svc.Create(context.Background(), up("invoice.pdf", []byte("<html><script>alert(1)</script></html>")))
	if !errors.Is(err, model.ErrUnsupportedType) {
		t.Errorf("got %v, want ErrUnsupportedType", err)
	}
	// A real PNG is accepted whatever it is called.
	if f, err := svc.Create(context.Background(), up("noext", png)); err != nil || f.ContentType != "image/png" {
		t.Errorf("png: %+v err=%v", f, err)
	}
}

func TestRejectsEmptyOversizedAndUnaddressedUploads(t *testing.T) {
	svc, meta, obj := newSvc(64)
	ctx := context.Background()

	if _, err := svc.Create(ctx, up("a.pdf", nil)); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("empty: got %v", err)
	}
	big := append(append([]byte{}, pdf...), bytes.Repeat([]byte("x"), 100)...)
	if _, err := svc.Create(ctx, up("a.pdf", big)); !errors.Is(err, model.ErrTooLarge) {
		t.Errorf("oversized: got %v", err)
	}
	in := up("a.pdf", pdf)
	in.ReceiverID = "  "
	if _, err := svc.Create(ctx, in); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("no receiver: got %v", err)
	}
	if len(meta.rows) != 0 || len(obj.data) != 0 {
		t.Error("a rejected upload must leave nothing behind")
	}
}

func TestFileNamesCannotCarryPathsOrControlCharacters(t *testing.T) {
	for in, want := range map[string]string{
		"../../etc/passwd":       "passwd",
		`C:\dir\scan.pdf`:        "scan.pdf",
		"a\r\nb.pdf":             "ab.pdf",
		"":                       "file",
		"..":                     "file",
		strings.Repeat("n", 400): strings.Repeat("n", 255),
	} {
		if got := cleanName(in); got != want {
			t.Errorf("cleanName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFailedMetadataInsertDoesNotOrphanTheObject(t *testing.T) {
	svc, meta, obj := newSvc(1 << 20)
	meta.failInsert = errors.New("mongo down")
	if _, err := svc.Create(context.Background(), up("a.pdf", pdf)); err == nil {
		t.Fatal("want error")
	}
	if len(obj.data) != 0 {
		t.Errorf("orphaned objects: %d", len(obj.data))
	}
}

func TestReplaceKeepsTheIDAndRemovesTheOldBytes(t *testing.T) {
	svc, _, obj := newSvc(1 << 20)
	ctx := context.Background()
	f, _ := svc.Create(ctx, up("a.pdf", pdf))

	g, err := svc.Replace(ctx, f.ID, up("b.png", png))
	if err != nil {
		t.Fatal(err)
	}
	if g.ID != f.ID || g.ObjectKey == f.ObjectKey || g.ContentType != "image/png" || g.FileName != "b.png" {
		t.Errorf("replaced = %+v", g)
	}
	if _, old := obj.data[f.ObjectKey]; old || len(obj.data) != 1 {
		t.Errorf("objects = %d, old still there: %v", len(obj.data), old)
	}
}

func TestFailedReplaceLeavesTheOriginalReadable(t *testing.T) {
	svc, meta, obj := newSvc(1 << 20)
	ctx := context.Background()
	f, _ := svc.Create(ctx, up("a.pdf", pdf))

	meta.failUpdate = errors.New("mongo down")
	if _, err := svc.Replace(ctx, f.ID, up("b.png", png)); err == nil {
		t.Fatal("want error")
	}
	d, err := svc.Open(ctx, f.ID)
	if err != nil {
		t.Fatalf("original unreadable: %v", err)
	}
	got, _ := io.ReadAll(d.Body)
	if !bytes.Equal(got, pdf) || len(obj.data) != 1 {
		t.Errorf("content changed or leaked: %d objects", len(obj.data))
	}
}

func TestDeleteRemovesBytesBeforeTheRecordAndCanBeRetried(t *testing.T) {
	svc, meta, obj := newSvc(1 << 20)
	ctx := context.Background()
	f, _ := svc.Create(ctx, up("a.pdf", pdf))

	obj.failDelete = model.ErrUpstream
	if err := svc.Delete(ctx, f.ID); !errors.Is(err, model.ErrUpstream) {
		t.Fatalf("got %v", err)
	}
	if _, ok := meta.rows[f.ID]; !ok {
		t.Fatal("record removed while the bytes are still there")
	}
	obj.failDelete = nil
	if err := svc.Delete(ctx, f.ID); err != nil || len(obj.data) != 0 || len(meta.rows) != 0 {
		t.Errorf("retry: err=%v objects=%d records=%d", err, len(obj.data), len(meta.rows))
	}
	if err := svc.Delete(ctx, f.ID); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("second delete: got %v", err)
	}
}

func TestListNeedsAReceiverAndClampsTheLimit(t *testing.T) {
	svc, _, _ := newSvc(1 << 20)
	if _, _, err := svc.List(context.Background(), "", 0, 0); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("got %v, want ErrInvalid", err)
	}
	_, _ = svc.Create(context.Background(), up("a.pdf", pdf))
	if files, total, err := svc.List(context.Background(), "patient:1", 100000, -5); err != nil || total != 1 || len(files) != 1 {
		t.Errorf("list: %v %d %v", files, total, err)
	}
}
