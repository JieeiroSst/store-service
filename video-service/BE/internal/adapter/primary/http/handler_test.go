package http_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/JIeeiroSst/video-service/config"
	httpadapter "github.com/JIeeiroSst/video-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/video-service/internal/adapter/secondary/buffer"
	"github.com/JIeeiroSst/video-service/internal/application"
	"github.com/JIeeiroSst/video-service/internal/domain/model"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
)

// ---- in-memory fakes for every driven port ----

type memStorage struct {
	mu   sync.Mutex
	objs map[string][]byte
}

func (m *memStorage) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) (int64, error) {
	b, err := io.ReadAll(r)
	m.mu.Lock()
	m.objs[key] = b
	m.mu.Unlock()
	return int64(len(b)), err
}
func (m *memStorage) ReadAt(_ context.Context, key string, size int64, p []byte, off int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return copy(p, m.objs[key][off:size]), nil
}
func (m *memStorage) Get(_ context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.objs[key]
	if !ok {
		return nil, port.ErrNotFound
	}
	return b, nil
}
func (m *memStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	b, err := m.Get(ctx, key)
	return io.NopCloser(bytes.NewReader(b)), err
}
func (m *memStorage) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.objs, key)
	m.mu.Unlock()
	return nil
}
func (m *memStorage) DeletePrefix(_ context.Context, prefix string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.objs {
		if strings.HasPrefix(k, prefix) {
			delete(m.objs, k)
		}
	}
	return nil
}
func (m *memStorage) has(prefix string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.objs {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

type memRepo struct {
	mu sync.Mutex
	m  map[string]model.Video
}

func (r *memRepo) Save(_ context.Context, v model.Video) error {
	r.mu.Lock()
	r.m[v.ID] = v
	r.mu.Unlock()
	return nil
}
func (r *memRepo) Get(_ context.Context, id string) (*model.Video, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.m[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &v, nil
}
func (r *memRepo) List(context.Context) ([]model.Video, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []model.Video
	for _, v := range r.m {
		out = append(out, v)
	}
	return out, nil
}
func (r *memRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	delete(r.m, id)
	r.mu.Unlock()
	return nil
}

type memViews struct {
	mu   sync.Mutex
	n    map[string]int64
	seen map[string]bool
}

func (v *memViews) Incr(_ context.Context, id, viewer string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if k := id + viewer; !v.seen[k] {
		v.seen[k] = true
		v.n[id]++
	}
	return nil
}
func (v *memViews) Counts(_ context.Context, ids []string) (map[string]int64, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := map[string]int64{}
	for _, id := range ids {
		out[id] = v.n[id]
	}
	return out, nil
}
func (v *memViews) Delete(_ context.Context, id string) error {
	v.mu.Lock()
	delete(v.n, id)
	v.mu.Unlock()
	return nil
}

type memQueue struct {
	mu  sync.Mutex
	ids []string
}

func (q *memQueue) Enqueue(_ context.Context, id string) error {
	q.mu.Lock()
	q.ids = append(q.ids, id)
	q.mu.Unlock()
	return nil
}
func (q *memQueue) Consume(context.Context, func(context.Context, string) error) error { return nil }

// fakeTranscoder writes what ffmpeg would: a master playlist, one rendition
// and a thumbnail.
type fakeTranscoder struct{}

func (fakeTranscoder) Transcode(_ context.Context, src, out string) (*port.TranscodeResult, error) {
	if b, _ := os.ReadFile(src); bytes.HasPrefix(b, []byte("BROKEN")) {
		return nil, fmt.Errorf("invalid data")
	}
	_ = os.MkdirAll(filepath.Join(out, "v0"), 0o755)
	_ = os.WriteFile(filepath.Join(out, "master.m3u8"), []byte("#EXTM3U\n"), 0o644)
	_ = os.WriteFile(filepath.Join(out, "v0", "index.m3u8"), []byte("#EXTM3U\n"), 0o644)
	_ = os.WriteFile(filepath.Join(out, "v0", "seg_0000.ts"), []byte("segment"), 0o644)
	_ = os.WriteFile(filepath.Join(out, "thumbnail.jpg"), []byte("jpeg"), 0o644)
	return &port.TranscodeResult{Duration: 12.5, Width: 1280, Height: 720}, nil
}

type env struct {
	srv       *httptest.Server
	storage   *memStorage
	queue     *memQueue
	transcode port.TranscodeUsecase
}

func newEnv(t *testing.T) *env {
	cfg := config.Load()
	cfg.Stream.ChunkSize = 1024
	raw := &memStorage{objs: map[string][]byte{}}
	storage := buffer.NewChunkStorage(raw, cfg)
	repo := buffer.NewCachedRepository(&memRepo{m: map[string]model.Video{}}, nil, cfg)
	views := &memViews{n: map[string]int64{}, seen: map[string]bool{}}
	queue := &memQueue{}

	uc := application.NewVideoService(repo, storage, views, queue)
	srv := httptest.NewServer(httpadapter.NewRouter(httpadapter.NewHandler(uc, cfg), cfg))
	t.Cleanup(srv.Close)
	return &env{srv, raw, queue, application.NewTranscodeService(repo, storage, fakeTranscoder{})}
}

func (e *env) upload(t *testing.T, title, desc string, data []byte) model.Video {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("title", title)
	_ = mw.WriteField("description", desc)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="demo.mp4"`)
	h.Set("Content-Type", "video/mp4")
	part, _ := mw.CreatePart(h)
	_, _ = part.Write(data)
	_ = mw.Close()

	resp, err := http.Post(e.srv.URL+"/api/videos", mw.FormDataContentType(), &body)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload: %v status=%v", err, resp)
	}
	var v model.Video
	mustJSON(t, resp.Body, &v)
	return v
}

func (e *env) get(path string, hdr ...string) *http.Response {
	req, _ := http.NewRequest("GET", e.srv.URL+path, nil)
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	resp, _ := http.DefaultClient.Do(req)
	return resp
}

func status(t *testing.T, resp *http.Response, want int, what string) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("%s: status=%d, want %d", what, resp.StatusCode, want)
	}
}

// ---- tests ----

func TestUploadStreamRangeDelete(t *testing.T) {
	e := newEnv(t)
	data := bytes.Repeat([]byte("0123456789abcdef"), 500) // 8000 bytes, spans several 1KiB chunks
	v := e.upload(t, "demo", "a description", data)

	if v.Status != model.StatusProcessing || v.Description != "a description" || len(e.queue.ids) != 1 || e.queue.ids[0] != v.ID {
		t.Fatalf("after upload: %+v queue=%v", v, e.queue.ids)
	}

	resp := e.get("/api/videos/" + v.ID + "/stream")
	got, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || !bytes.Equal(got, data) || resp.Header.Get("Accept-Ranges") != "bytes" {
		t.Fatalf("full stream: status=%d len=%d", resp.StatusCode, len(got))
	}

	resp = e.get("/api/videos/"+v.ID+"/stream", "Range", "bytes=1000-3000")
	got, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusPartialContent || !bytes.Equal(got, data[1000:3001]) ||
		resp.Header.Get("Content-Range") != "bytes 1000-3000/8000" {
		t.Fatalf("range: status=%d cr=%q len=%d", resp.StatusCode, resp.Header.Get("Content-Range"), len(got))
	}

	etag := resp.Header.Get("ETag")
	status(t, e.get("/api/videos/"+v.ID+"/stream", "If-None-Match", etag), http.StatusNotModified, "if-none-match")

	req, _ := http.NewRequest("DELETE", e.srv.URL+"/api/videos/"+v.ID, nil)
	resp, _ = http.DefaultClient.Do(req)
	status(t, resp, http.StatusNoContent, "delete")
	status(t, e.get("/api/videos/"+v.ID+"/stream"), http.StatusNotFound, "stream after delete")
	if e.storage.has("videos/") {
		t.Fatal("original left behind after delete")
	}
}

func TestHLSAndThumbnailOnlyOnceReady(t *testing.T) {
	e := newEnv(t)
	v := e.upload(t, "clip", "", []byte("some video bytes"))

	status(t, e.get("/api/videos/"+v.ID+"/hls/master.m3u8"), http.StatusNotFound, "hls while processing")
	status(t, e.get("/api/videos/"+v.ID+"/thumbnail"), http.StatusNotFound, "thumbnail while processing")

	if err := e.transcode.Process(context.Background(), v.ID); err != nil {
		t.Fatal(err)
	}
	if err := e.transcode.Process(context.Background(), v.ID); err != nil { // redelivery is a no-op
		t.Fatal(err)
	}

	resp := e.get("/api/videos/" + v.ID)
	var got model.Video
	mustJSON(t, resp.Body, &got)
	if got.Status != model.StatusReady || got.Duration != 12.5 || got.Height != 720 {
		t.Fatalf("after transcode: %+v", got)
	}

	resp = e.get("/api/videos/" + v.ID + "/hls/master.m3u8")
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "application/vnd.apple.mpegurl" || string(body) != "#EXTM3U\n" {
		t.Fatalf("master: status=%d ct=%q", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	resp = e.get("/api/videos/" + v.ID + "/hls/v0/seg_0000.ts")
	if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "video/mp2t" {
		t.Fatalf("segment: status=%d ct=%q", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	status(t, e.get("/api/videos/"+v.ID+"/thumbnail"), 200, "thumbnail")

	// only files of the HLS package are reachable
	for _, bad := range []string{"../videos/x", "v0/../../x", "secret.txt", "v0/seg_1.ts"} {
		status(t, e.get("/api/videos/"+v.ID+"/hls/"+bad), http.StatusNotFound, "hls path "+bad)
	}
}

func TestUndecodableVideoIsMarkedFailed(t *testing.T) {
	e := newEnv(t)
	v := e.upload(t, "bad", "", []byte("BROKEN file"))

	if err := e.transcode.Process(context.Background(), v.ID); err != nil {
		t.Fatalf("permanent failure must not be retried: %v", err)
	}
	var got model.Video
	mustJSON(t, e.get("/api/videos/"+v.ID).Body, &got)
	if got.Status != model.StatusFailed {
		t.Fatalf("status = %s", got.Status)
	}
	var list struct{ Total int }
	mustJSON(t, e.get("/api/videos").Body, &list)
	if list.Total != 0 {
		t.Fatal("failed video should be hidden from the list")
	}
}

func TestSearchSortViewsRelated(t *testing.T) {
	e := newEnv(t)
	a := e.upload(t, "Go concurrency", "goroutines and channels", []byte("a"))
	b := e.upload(t, "Kubernetes intro", "pods", []byte("b"))
	c := e.upload(t, "Rust ownership", "borrow checker", []byte("c"))
	for _, v := range []model.Video{a, b, c} {
		_ = e.transcode.Process(context.Background(), v.ID)
	}

	// views count once per viewer
	for i := 0; i < 3; i++ {
		resp, _ := http.Post(e.srv.URL+"/api/videos/"+b.ID+"/view", "", nil)
		status(t, resp, http.StatusNoContent, "view")
	}
	var list struct {
		Items []model.Video
		Total int
	}
	mustJSON(t, e.get("/api/videos?sort=popular").Body, &list)
	if list.Total != 3 || list.Items[0].ID != b.ID || list.Items[0].Views != 1 {
		t.Fatalf("popular: %+v", list)
	}

	list.Items = nil
	mustJSON(t, e.get("/api/videos?q=CHANNELS").Body, &list) // case-insensitive, searches description
	if list.Total != 1 || list.Items[0].ID != a.ID {
		t.Fatalf("search: %+v", list)
	}

	var rel struct{ Items []model.Video }
	mustJSON(t, e.get("/api/videos/"+b.ID+"/related").Body, &rel)
	if len(rel.Items) != 2 {
		t.Fatalf("related: %+v", rel)
	}
	for _, v := range rel.Items {
		if v.ID == b.ID {
			t.Fatal("related contains the video itself")
		}
	}

	status(t, mustPost(e.srv.URL+"/api/videos/nope/view"), http.StatusNotFound, "view unknown video")
}

func TestRejectsNonVideoAndBadID(t *testing.T) {
	e := newEnv(t)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", "x.txt") // application/octet-stream
	_, _ = part.Write([]byte("hi"))
	_ = mw.Close()
	resp, _ := http.Post(e.srv.URL+"/api/videos", mw.FormDataContentType(), &body)
	status(t, resp, http.StatusBadRequest, "non-video upload")
	status(t, e.get("/api/videos/..%2f..%2fetc/stream"), http.StatusNotFound, "bad id")
}

func mustPost(url string) *http.Response {
	resp, _ := http.Post(url, "", nil)
	return resp
}
