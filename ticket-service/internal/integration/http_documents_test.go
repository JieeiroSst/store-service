package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeUploadService behaves like upload-service with a service key: it refuses anything without it and keeps files by receiver.
type fakeUploadService struct {
	*httptest.Server
	mu    sync.Mutex
	files map[string]uploaded
	n     int
	bad   int // calls refused for a missing or wrong key
}

type uploaded struct {
	receiver string
	name     string
	data     []byte
}

func newFakeUploadService(t *testing.T, key string) *fakeUploadService {
	f := &fakeUploadService{files: map[string]uploaded{}}
	mux := http.NewServeMux()
	guard := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Service-Key") != key {
				f.mu.Lock()
				f.bad++
				f.mu.Unlock()
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("POST /api/v1/upload", guard(func(w http.ResponseWriter, r *http.Request) {
		file, hdr, err := r.FormFile("file")
		if err != nil || r.URL.Query().Get("receiver_id") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data, _ := io.ReadAll(file)
		f.mu.Lock()
		f.n++
		id := fmt.Sprintf("u-%d", f.n)
		f.files[id] = uploaded{receiver: r.URL.Query().Get("receiver_id"), name: hdr.Filename, data: data}
		f.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
	}))
	mux.HandleFunc("GET /api/v1/upload/{id}/content", guard(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		x, ok := f.files[r.PathValue("id")]
		f.mu.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprint(len(x.data)))
		_, _ = w.Write(x.data)
	}))
	mux.HandleFunc("DELETE /api/v1/upload/{id}", guard(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		delete(f.files, r.PathValue("id"))
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	f.Server = httptest.NewServer(mux)
	t.Cleanup(f.Close)
	return f
}

func (f *fakeUploadService) of(receiver string) []uploaded {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []uploaded
	for _, x := range f.files {
		if x.receiver == receiver {
			out = append(out, x)
		}
	}
	return out
}

func waitFor(t *testing.T, what string, d time.Duration, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(d); time.Now().Before(deadline); time.Sleep(150 * time.Millisecond) {
		if ok() {
			return
		}
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestHTTPDocuments(t *testing.T) {
	up := newFakeUploadService(t, "upload-key")
	a := startApp(t, map[string]string{"UploadServiceURL": up.URL, "UploadServiceKey": "upload-key", "SweepIntervalSeconds": "1", "InternalAPIKey": "cluster-secret"})
	const org = int64(1_000_000_002)
	eid, tid := publishedEvent(a, 10)
	paid := buy(a, 5, eid, tid, 2)
	oid := int64(paid["id"].(float64))
	op := fmt.Sprintf("/api/v1/orders/%d", oid)

	// the sweeper makes the documents and keeps them under the buyer's id
	waitFor(t, "the documents", 10*time.Second, func() bool { return len(up.of("ticket-user:5")) == 2 })
	if len(up.of("ticket-user:6")) != 0 {
		t.Fatal("files under somebody else's id")
	}
	if up.bad != 0 {
		t.Fatalf("%d calls to upload-service went without the right service key", up.bad)
	}
	for _, f := range up.of("ticket-user:5") {
		if !isPDF(f.data) || !strings.HasSuffix(f.name, ".pdf") {
			t.Fatalf("stored %q is not a PDF", f.name)
		}
	}
	docs := a.must(200, "GET", op+"/documents", 5, nil)["items"].([]any)
	if len(docs) != 2 {
		t.Fatalf("documents: %v", docs)
	}
	a.must(404, "GET", op+"/documents", 6, nil)
	a.must(401, "GET", op+"/documents", 0, nil)

	// downloads go through this service, which checks whose they are
	st, hdr, body := a.raw("GET", op+"/documents/invoice", 5, nil)
	if st != 200 || hdr.Get("Content-Type") != "application/pdf" || !strings.HasPrefix(hdr.Get("Content-Disposition"), "attachment") || !isPDF([]byte(body)) {
		t.Fatalf("download: %d %v", st, hdr)
	}
	if hdr.Get("X-Content-Type-Options") != "nosniff" || hdr.Get("Cache-Control") != "private, no-store" {
		t.Fatalf("headers: %v", hdr)
	}
	if st, _, _ := a.raw("GET", op+"/documents/invoice", 6, nil); st != 404 {
		t.Fatalf("a stranger downloading: %d", st)
	}
	if st, _, _ := a.raw("GET", op+"/documents/passport", 5, nil); st != 400 {
		t.Fatalf("an unknown kind: %d", st)
	}
	if st, _, _ := a.raw("GET", op+"/documents/tickets", org, nil); st != 404 && st != 403 && st != 200 {
		t.Fatalf("the organizer: %d", st)
	}

	// PDFs drawn on request
	st, hdr, body = a.raw("GET", op+"/invoice?format=pdf", 5, nil)
	if st != 200 || hdr.Get("Content-Type") != "application/pdf" || !strings.HasPrefix(hdr.Get("Content-Disposition"), "inline") || !isPDF([]byte(body)) {
		t.Fatalf("invoice pdf: %d %v", st, hdr)
	}
	tk := paid["tickets"].([]any)[0].(map[string]any)
	tp := fmt.Sprintf("/api/v1/tickets/%d/pdf", int64(tk["id"].(float64)))
	st, hdr, body = a.raw("GET", tp, 5, nil)
	if st != 200 || !isPDF([]byte(body)) || !strings.HasPrefix(hdr.Get("Content-Disposition"), "attachment") {
		t.Fatalf("ticket pdf: %d %v", st, hdr)
	}
	if st, _, _ := a.raw("GET", tp, 6, nil); st != 404 {
		t.Fatalf("a stranger's ticket pdf: %d", st)
	}
	if st, _, _ := a.raw("GET", tp, 0, nil); st != 401 {
		t.Fatalf("an anonymous ticket pdf: %d", st)
	}

	// the same through the cluster API
	if st, _, body := a.raw("GET", "/internal/v1"+strings.TrimPrefix(op, "/api/v1")+"/documents/tickets", 0, map[string]string{"X-Internal-Key": "cluster-secret"}); st != 200 || !isPDF([]byte(body)) {
		t.Fatalf("internal download: %d", st)
	}
	_ = bytes.MinRead
}

func TestDocumentsAreNotStoredWithoutAnUploadService(t *testing.T) {
	a := startApp(t, map[string]string{"SweepIntervalSeconds": "1"})
	eid, tid := publishedEvent(a, 5)
	paid := buy(a, 5, eid, tid, 1)
	op := fmt.Sprintf("/api/v1/orders/%d", int64(paid["id"].(float64)))
	time.Sleep(2500 * time.Millisecond)
	if docs := a.must(200, "GET", op+"/documents", 5, nil)["items"].([]any); len(docs) != 0 {
		t.Fatalf("documents without a store: %v", docs)
	}
	if st, _, body := a.raw("GET", op+"/invoice?format=pdf", 5, nil); st != 200 || !isPDF([]byte(body)) {
		t.Fatalf("an invoice is still drawn on request: %d", st)
	}
	if st, _, _ := a.raw("GET", op+"/documents/invoice", 5, nil); st != 404 {
		t.Fatalf("nothing stored: %d", st)
	}
}
