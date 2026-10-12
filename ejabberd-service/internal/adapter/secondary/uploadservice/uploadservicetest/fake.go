package uploadservicetest

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type File struct {
	ID          string
	ReceiverID  string
	FileName    string
	ContentType string
	Data        []byte
}

type Fake struct {
	ServiceKey string
	MaxBytes   int64
	Allowed    []string
	Uploads    atomic.Int64
	mu         sync.Mutex
	files      []File
}

func NewFake(serviceKey string) *Fake {
	return &Fake{
		ServiceKey: serviceKey,
		MaxBytes:   20 << 20,
		Allowed:    []string{"application/pdf", "image/jpeg", "image/png", "image/webp", "image/gif"},
	}
}

func (f *Fake) Files() []File {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.files)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (f *Fake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Service-Key") != f.ServiceKey {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/upload":
		f.create(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/upload":
		receiver := r.URL.Query().Get("receiver_id")
		var items []map[string]any
		for _, file := range f.Files() {
			if file.ReceiverID == receiver {
				items = append(items, map[string]any{
					"id": file.ID, "receiver_id": file.ReceiverID, "file_name": file.FileName,
					"content_type": file.ContentType, "size": len(file.Data),
					"url": "/api/v1/upload/" + file.ID + "/content",
				})
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"files": items, "total_count": len(items)})
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/content"):
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/upload/"), "/content")
		for _, file := range f.Files() {
			if file.ID == id {
				w.Header().Set("Content-Type", file.ContentType)
				w.Header().Set("Content-Length", strconv.Itoa(len(file.Data)))
				w.Header().Set("Content-Disposition", `attachment; filename="`+file.FileName+`"`)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(file.Data)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}
}

func (f *Fake) create(w http.ResponseWriter, r *http.Request) {
	receiver := strings.TrimSpace(r.URL.Query().Get("receiver_id"))
	if receiver == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid argument: receiver_id is required"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid argument: multipart form field 'file' is required"})
		return
	}
	defer file.Close()
	data, _ := io.ReadAll(io.LimitReader(file, f.MaxBytes+1))
	if int64(len(data)) > f.MaxBytes {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "file too large"})
		return
	}
	ct := http.DetectContentType(data)
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	if !slices.Contains(f.Allowed, ct) {
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "unsupported file type: " + ct})
		return
	}
	f.Uploads.Add(1)
	f.mu.Lock()
	id := "file-" + strconv.Itoa(len(f.files)+1)
	f.files = append(f.files, File{ID: id, ReceiverID: receiver, FileName: header.Filename, ContentType: ct, Data: data})
	f.mu.Unlock()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "receiver_id": receiver, "file_name": header.Filename, "content_type": ct, "size": len(data)})
}
