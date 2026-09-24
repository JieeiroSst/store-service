package document

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

// fakeUpload mimics upload-service: files keyed by id, each owned by a receiver.
type fakeUpload struct {
	auth    []string
	deleted []string
	files   map[string]map[string]any
	status  int // when set, every call answers with it
}

func (f *fakeUpload) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f.auth = append(f.auth, r.Header.Get("Authorization"))
		if f.status != 0 {
			w.WriteHeader(f.status)
			_, _ = w.Write([]byte(`{"error":"nope"}`))
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/upload")
		reply := func(code int, v any) { w.WriteHeader(code); _ = json.NewEncoder(w).Encode(v) }
		switch {
		case r.Method == "POST" && path == "":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Errorf("not multipart: %v", err)
			}
			_, hdr, err := r.FormFile("file")
			if err != nil {
				t.Errorf("no file field: %v", err)
				return
			}
			rec := map[string]any{"id": "f-new", "receiver_id": r.URL.Query().Get("receiver_id"), "file_name": hdr.Filename,
				"content_type": "application/pdf", "size": hdr.Size, "created_at": time.Unix(1_700_000_000, 0).UTC()}
			f.files["f-new"] = rec
			reply(201, rec)
		case r.Method == "GET" && path == "":
			var out []map[string]any
			for _, rec := range f.files {
				if rec["receiver_id"] == r.URL.Query().Get("receiver_id") {
					out = append(out, rec)
				}
			}
			reply(200, map[string]any{"files": out, "total_count": len(out)})
		case r.Method == "GET" && strings.HasSuffix(path, "/content"):
			_, _ = w.Write([]byte("%PDF-bytes"))
		case r.Method == "GET":
			if rec, ok := f.files[strings.TrimPrefix(path, "/")]; ok {
				reply(200, rec)
				return
			}
			reply(404, map[string]string{"error": "not found"})
		case r.Method == "DELETE":
			f.deleted = append(f.deleted, strings.TrimPrefix(path, "/"))
			w.WriteHeader(204)
		}
	}
}

func setup(t *testing.T) (*Client, *fakeUpload) {
	f := &fakeUpload{files: map[string]map[string]any{
		"mine":  {"id": "mine", "receiver_id": "patient:1", "file_name": "a.pdf", "content_type": "application/pdf", "size": 10},
		"other": {"id": "other", "receiver_id": "patient:2", "file_name": "b.pdf", "content_type": "application/pdf", "size": 10},
	}}
	srv := httptest.NewServer(f.handler(t))
	t.Cleanup(srv.Close)
	return New(&config.Config{Upload: config.UpstreamConfig{BaseURL: srv.URL, Timeout: time.Second}}), f
}

func TestUploadTagsTheFileWithThePatientAndForwardsTheCallersToken(t *testing.T) {
	c, f := setup(t)
	ctx := port.WithBearer(context.Background(), "tok-1")

	d, err := c.Upload(ctx, 7, port.DocumentUpload{FileName: "scan.pdf", Body: strings.NewReader("%PDF-x")})
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "f-new" || d.PatientID != 7 || d.FileName != "scan.pdf" {
		t.Errorf("doc = %+v", d)
	}
	if f.files["f-new"]["receiver_id"] != "patient:7" {
		t.Errorf("receiver = %v", f.files["f-new"]["receiver_id"])
	}
	if f.auth[0] != "Bearer tok-1" {
		t.Errorf("Authorization = %q", f.auth[0])
	}
}

func TestListIsScopedToThePatient(t *testing.T) {
	c, _ := setup(t)
	docs, total, err := c.List(context.Background(), 1, 20, 0)
	if err != nil || total != 1 || len(docs) != 1 || docs[0].ID != "mine" {
		t.Errorf("docs=%+v total=%d err=%v", docs, total, err)
	}
}

// The heart of the design: knowing another patient's file id must not let you
// read or delete it by going through your own patient.
func TestAnotherPatientsFileIsNotFound(t *testing.T) {
	c, f := setup(t)
	ctx := context.Background()

	if _, err := c.Open(ctx, 1, "other"); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("open: got %v, want ErrNotFound", err)
	}
	if err := c.Delete(ctx, 1, "other"); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("delete: got %v, want ErrNotFound", err)
	}
	if len(f.deleted) != 0 {
		t.Errorf("deleted another patient's file: %v", f.deleted)
	}

	d, err := c.Open(ctx, 1, "mine")
	if err != nil {
		t.Fatal(err)
	}
	defer d.Body.Close()
	if b, _ := io.ReadAll(d.Body); !bytes.Equal(b, []byte("%PDF-bytes")) {
		t.Errorf("content = %q", b)
	}
	if err := c.Delete(ctx, 1, "mine"); err != nil || len(f.deleted) != 1 {
		t.Errorf("delete own: err=%v deleted=%v", err, f.deleted)
	}
}

func TestUploadServiceAnswersAreClassified(t *testing.T) {
	cases := map[string]struct {
		status int
		want   error
	}{
		"too large":     {413, model.ErrInvalid},
		"wrong type":    {415, model.ErrInvalid},
		"bad request":   {400, model.ErrInvalid},
		"not found":     {404, model.ErrNotFound},
		"we are denied": {401, model.ErrUpstream},
		"it is down":    {500, model.ErrUpstream},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c, f := setup(t)
			f.status = tc.status
			_, err := c.Upload(context.Background(), 1, port.DocumentUpload{FileName: "a.pdf", Body: strings.NewReader("x")})
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDisabledWithoutABaseURL(t *testing.T) {
	if New(&config.Config{}).Enabled() {
		t.Error("client with no base URL must be disabled")
	}
}
