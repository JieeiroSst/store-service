package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"strings"
	"testing"

	"github.com/JIeeiroSst/draw-image-service/config"
	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"github.com/google/uuid"
)

type namedImage struct {
	image.Image
	name string
}

type fakeProcessor struct {
	columns int
	failOn  string
	center  string
	others  []string
	shrunk  [][2]int
}

func (f *fakeProcessor) Decode(r io.Reader) (image.Image, error) {
	b, _ := io.ReadAll(r)
	if string(b) == f.failOn {
		return nil, domain.ErrInvalidImage
	}
	return namedImage{image.NewNRGBA(image.Rect(0, 0, 1, 1)), string(b)}, nil
}

func (f *fakeProcessor) Thumbnail(_ image.Image, w, h int) image.Image {
	return image.NewNRGBA(image.Rect(0, 0, w, h))
}

func (f *fakeProcessor) Shrink(img image.Image, w, h int) image.Image {
	f.shrunk = append(f.shrunk, [2]int{w, h})
	return img
}

func (f *fakeProcessor) ComposeCenter(center image.Image, others []image.Image, w, h int) image.Image {
	f.center = center.(namedImage).name
	f.others = nil
	for _, o := range others {
		f.others = append(f.others, o.(namedImage).name)
	}
	return image.NewNRGBA(image.Rect(0, 0, w, h))
}

func (f *fakeProcessor) Compose(cells []image.Image, w, h, columns int) image.Image {
	f.columns = columns
	rows := (len(cells) + columns - 1) / columns
	return image.NewNRGBA(image.Rect(0, 0, w*columns, h*rows))
}

func (f *fakeProcessor) EncodeJPEG(w io.Writer, _ image.Image, _ int) error {
	_, err := w.Write([]byte("jpeg"))
	return err
}

type fakeStorage struct {
	objects map[string][]byte
	putErr  error
}

func (s *fakeStorage) Bucket() string { return "test-bucket" }

func (s *fakeStorage) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	if s.putErr != nil {
		return s.putErr
	}
	b, _ := io.ReadAll(r)
	s.objects[key] = b
	return nil
}

func (s *fakeStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	b, ok := s.objects[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (s *fakeStorage) Delete(_ context.Context, key string) error {
	delete(s.objects, key)
	return nil
}

type fakeRepo struct {
	rows    map[string]domain.Collage
	saveErr error
}

func (r *fakeRepo) Save(_ context.Context, c *domain.Collage) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.rows[c.ID] = *c
	return nil
}

func (r *fakeRepo) FindByID(_ context.Context, id string) (*domain.Collage, error) {
	c, ok := r.rows[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &c, nil
}

type fixture struct {
	svc       *collageService
	processor *fakeProcessor
	storage   *fakeStorage
	repo      *fakeRepo
}

func newFixture() *fixture {
	f := &fixture{
		processor: &fakeProcessor{failOn: "bad"},
		storage:   &fakeStorage{objects: map[string][]byte{}},
		repo:      &fakeRepo{rows: map[string]domain.Collage{}},
	}
	cfg := &config.Config{Collage: config.CollageConfig{
		CellWidth: 10, CellHeight: 10, Columns: 2, MaxImages: 3, JPEGQuality: 90,
		CenterWidth: 80, CenterHeight: 60,
	}}
	f.svc = NewCollageService(f.processor, f.storage, f.repo, cfg).(*collageService)
	return f
}

func sources(contents ...string) []domain.Source {
	out := make([]domain.Source, len(contents))
	for i, c := range contents {
		out[i] = domain.Source{Name: c, Content: strings.NewReader(c)}
	}
	return out
}

func TestCreateStoresObjectAndRecord(t *testing.T) {
	f := newFixture()
	got, err := f.svc.Create(context.Background(), domain.CollageRequest{
		Sources: sources("a", "b", "c"), Layout: domain.LayoutGrid,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(got.ID); err != nil {
		t.Fatalf("id %q is not a uuid", got.ID)
	}
	if f.processor.columns != 2 || got.Width != 20 || got.Height != 20 || got.ImageCount != 3 {
		t.Fatalf("columns=%d size=%dx%d count=%d", f.processor.columns, got.Width, got.Height, got.ImageCount)
	}
	if got.Bucket != "test-bucket" || got.ObjectKey != "collages/"+got.ID+".jpg" || got.Size != 4 {
		t.Fatalf("unexpected location %+v", got)
	}
	if string(f.storage.objects[got.ObjectKey]) != "jpeg" {
		t.Fatalf("object not stored")
	}
	if _, ok := f.repo.rows[got.ID]; !ok {
		t.Fatalf("record not saved")
	}
}

func TestCreateGivesEachCollageItsOwnID(t *testing.T) {
	f := newFixture()
	a, _ := f.svc.Create(context.Background(), domain.CollageRequest{Sources: sources("a")})
	b, _ := f.svc.Create(context.Background(), domain.CollageRequest{Sources: sources("a")})
	if a.ID == b.ID || len(f.storage.objects) != 2 {
		t.Fatalf("ids collide: %s %s", a.ID, b.ID)
	}
}

func TestCreateRowPutsAllOnOneLine(t *testing.T) {
	got, err := newFixture().svc.Create(context.Background(), domain.CollageRequest{
		Sources: sources("a", "b", "c"), Layout: domain.LayoutRow,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Width != 30 || got.Height != 10 || got.Layout != domain.LayoutRow {
		t.Fatalf("size=%dx%d layout=%s", got.Width, got.Height, got.Layout)
	}
}

func TestCreateCenterFeaturesFirstSource(t *testing.T) {
	f := newFixture()
	got, err := f.svc.Create(context.Background(), domain.CollageRequest{
		Sources: sources("chair", "sale", "logo"), Layout: domain.LayoutCenter,
	})
	if err != nil {
		t.Fatal(err)
	}
	if f.processor.center != "chair" || strings.Join(f.processor.others, ",") != "sale,logo" {
		t.Fatalf("center=%q others=%v", f.processor.center, f.processor.others)
	}
	if got.Layout != domain.LayoutCenter || got.Width != 80 || got.Height != 60 || got.Columns != 0 {
		t.Fatalf("got %+v", got)
	}
	if len(f.processor.shrunk) != 3 || f.processor.shrunk[0] != [2]int{80, 60} {
		t.Fatalf("every decoded image must be shrunk to the canvas, got %v", f.processor.shrunk)
	}
}

func TestCreateCenterPassesEveryOtherImage(t *testing.T) {
	f := newFixture()
	f.svc.cfg.MaxImages = 30
	names := make([]string, 30)
	for i := range names {
		names[i] = fmt.Sprintf("img%d", i)
	}
	if _, err := f.svc.Create(context.Background(), domain.CollageRequest{
		Sources: sources(names...), Layout: domain.LayoutCenter,
	}); err != nil {
		t.Fatal(err)
	}
	if f.processor.center != "img0" || len(f.processor.others) != 29 || f.processor.others[28] != "img29" {
		t.Fatalf("center=%q others=%v", f.processor.center, f.processor.others)
	}
}

func TestCreateClampsColumnsToImageCount(t *testing.T) {
	f := newFixture()
	if _, err := f.svc.Create(context.Background(), domain.CollageRequest{
		Sources: sources("a"), Layout: domain.LayoutGrid, Columns: 5,
	}); err != nil {
		t.Fatal(err)
	}
	if f.processor.columns != 1 {
		t.Fatalf("columns=%d, want 1", f.processor.columns)
	}
}

func TestCreateRemovesObjectWhenRecordFails(t *testing.T) {
	f := newFixture()
	f.repo.saveErr = errors.New("db down")
	if _, err := f.svc.Create(context.Background(), domain.CollageRequest{Sources: sources("a")}); err == nil {
		t.Fatal("expected error")
	}
	if len(f.storage.objects) != 0 {
		t.Fatalf("orphan object left behind: %v", f.storage.objects)
	}
}

func TestCreateDoesNotSaveRecordWhenUploadFails(t *testing.T) {
	f := newFixture()
	f.storage.putErr = errors.New("minio down")
	if _, err := f.svc.Create(context.Background(), domain.CollageRequest{Sources: sources("a")}); err == nil {
		t.Fatal("expected error")
	}
	if len(f.repo.rows) != 0 {
		t.Fatalf("record saved without object")
	}
}

func TestCreateErrors(t *testing.T) {
	tests := []struct {
		name string
		req  domain.CollageRequest
		want error
	}{
		{"no images", domain.CollageRequest{}, domain.ErrNoImages},
		{"too many", domain.CollageRequest{Sources: sources("a", "b", "c", "d")}, domain.ErrTooManyImages},
		{"bad layout", domain.CollageRequest{Sources: sources("a"), Layout: "zigzag"}, domain.ErrInvalidLayout},
		{"bad image", domain.CollageRequest{Sources: sources("a", "bad")}, domain.ErrInvalidImage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newFixture().svc.Create(context.Background(), tt.req)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err=%v, want %v", err, tt.want)
			}
		})
	}
}

func TestOpenReturnsStoredImage(t *testing.T) {
	f := newFixture()
	created, err := f.svc.Create(context.Background(), domain.CollageRequest{Sources: sources("a")})
	if err != nil {
		t.Fatal(err)
	}

	got, rc, err := f.svc.Open(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	if got.ID != created.ID || string(data) != "jpeg" {
		t.Fatalf("got %+v data=%q", got, data)
	}
}

func TestGetValidatesID(t *testing.T) {
	f := newFixture()
	if _, err := f.svc.Get(context.Background(), "not-a-uuid"); !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("err=%v, want ErrInvalidID", err)
	}
	if _, err := f.svc.Get(context.Background(), uuid.NewString()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
}
