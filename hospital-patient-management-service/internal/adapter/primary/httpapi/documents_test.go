package httpapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type fakeDocs struct {
	err      error
	uploaded []byte
	name     string
	deleted  string
}

func (f *fakeDocs) Upload(_ context.Context, id int32, in port.DocumentUpload) (*model.Document, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.uploaded, _ = io.ReadAll(in.Body)
	f.name = in.FileName
	return &model.Document{ID: "d1", PatientID: id, FileName: in.FileName, ContentType: "application/pdf", Size: int64(len(f.uploaded))}, nil
}
func (f *fakeDocs) List(context.Context, int32, int, int) ([]model.Document, int64, error) {
	return []model.Document{{ID: "d1", PatientID: 1, FileName: "a.pdf"}}, 1, f.err
}
func (f *fakeDocs) Open(context.Context, int32, string) (*port.DocumentDownload, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &port.DocumentDownload{
		Document: &model.Document{ID: "d1", FileName: `re"port.pdf`, ContentType: "application/pdf", Size: 5},
		Body:     io.NopCloser(strings.NewReader("%PDF-")),
	}, nil
}
func (f *fakeDocs) Delete(_ context.Context, _ int32, doc string) error {
	f.deleted = doc
	return f.err
}

func serveDocs(uc *fakeDocs) http.Handler {
	mux := http.NewServeMux()
	NewDocumentHandler(uc).Register(mux)
	return mux
}

func TestUploadForwardsTheFile(t *testing.T) {
	uc := &fakeDocs{}
	body, ct := multipartBody(t, map[string][]string{"file": {"%PDF-1"}})
	rec := do(serveDocs(uc), http.MethodPost, "/v1/patients/3/documents", body, ct)
	if rec.Code != 201 || string(uc.uploaded) != "%PDF-1" || !strings.Contains(rec.Body.String(), `"patient_id":3`) {
		t.Errorf("%d %s uploaded=%q", rec.Code, rec.Body, uc.uploaded)
	}
	body, ct = multipartBody(t, map[string][]string{"other": {"x"}})
	if rec := do(serveDocs(uc), http.MethodPost, "/v1/patients/3/documents", body, ct); rec.Code != 400 {
		t.Errorf("missing file field: %d", rec.Code)
	}
}

func TestDownloadIsAnAttachmentWithSafeHeaders(t *testing.T) {
	rec := do(serveDocs(&fakeDocs{}), http.MethodGet, "/v1/patients/1/documents/d1/content", nil, "")
	h := rec.Header()
	if rec.Code != 200 || !bytes.Equal(rec.Body.Bytes(), []byte("%PDF-")) {
		t.Fatalf("%d %q", rec.Code, rec.Body)
	}
	if h.Get("Content-Type") != "application/pdf" || h.Get("X-Content-Type-Options") != "nosniff" ||
		h.Get("Cache-Control") != "private, no-store" || !strings.HasPrefix(h.Get("Content-Disposition"), "attachment") {
		t.Errorf("headers = %v", h)
	}
	if strings.Contains(h.Get("Content-Disposition"), `re"port`) {
		t.Errorf("unescaped file name in %q", h.Get("Content-Disposition"))
	}
}

func TestListDeleteAndErrors(t *testing.T) {
	uc := &fakeDocs{}
	if rec := do(serveDocs(uc), http.MethodGet, "/v1/patients/1/documents", nil, ""); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"total_count":1`) {
		t.Errorf("list: %d %s", rec.Code, rec.Body)
	}
	if rec := do(serveDocs(uc), http.MethodDelete, "/v1/patients/1/documents/d9", nil, ""); rec.Code != 204 || uc.deleted != "d9" {
		t.Errorf("delete: %d deleted=%q", rec.Code, uc.deleted)
	}
	for err, code := range map[error]int{model.ErrNotFound: 404, model.Invalid("too big"): 400, model.ErrUpstream: 503} {
		if rec := do(serveDocs(&fakeDocs{err: err}), http.MethodGet, "/v1/patients/1/documents/d1/content", nil, ""); rec.Code != code {
			t.Errorf("%v: %d, want %d", err, rec.Code, code)
		}
	}
}
