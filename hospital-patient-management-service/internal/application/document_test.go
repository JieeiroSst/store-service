package application

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type fakeDocs struct {
	off      bool
	count    int64
	uploaded int
}

func (f *fakeDocs) Enabled() bool { return !f.off }
func (f *fakeDocs) Upload(_ context.Context, id int32, in port.DocumentUpload) (*model.Document, error) {
	f.uploaded++
	return &model.Document{ID: "d1", PatientID: id, FileName: in.FileName}, nil
}
func (f *fakeDocs) List(context.Context, int32, int, int) ([]model.Document, int64, error) {
	return nil, f.count, nil
}
func (f *fakeDocs) Open(context.Context, int32, string) (*port.DocumentDownload, error) {
	return &port.DocumentDownload{Document: &model.Document{}, Body: io.NopCloser(strings.NewReader(""))}, nil
}
func (f *fakeDocs) Delete(context.Context, int32, string) error { return nil }

type missingPatients struct{ fakePatients }

func (missingPatients) Get(context.Context, int32) (*model.Patient, error) {
	return nil, model.ErrNotFound
}

func TestDocumentsRequireAnExistingPatientAndAConfiguredService(t *testing.T) {
	ctx := context.Background()
	in := port.DocumentUpload{FileName: "a.pdf", Body: strings.NewReader("x")}

	gw := &fakeDocs{}
	if _, err := NewDocumentService(missingPatients{}, gw).Upload(ctx, 9, in); !errors.Is(err, model.ErrNotFound) || gw.uploaded != 0 {
		t.Errorf("unknown patient: err=%v uploaded=%d", err, gw.uploaded)
	}
	if _, err := NewDocumentService(fakePatients{}, &fakeDocs{off: true}).Upload(ctx, 1, in); !errors.Is(err, model.ErrUpstream) {
		t.Errorf("disabled: got %v, want ErrUpstream", err)
	}
	if _, err := NewDocumentService(fakePatients{}, gw).Upload(ctx, 0, in); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("bad id: got %v, want ErrInvalid", err)
	}
	if d, err := NewDocumentService(fakePatients{}, gw).Upload(ctx, 1, in); err != nil || d.PatientID != 1 {
		t.Errorf("happy path: %+v err=%v", d, err)
	}
}

func TestPatientWithDocumentsCannotBeDeleted(t *testing.T) {
	ctx := context.Background()
	repo := &memPatientRepo{}
	// memPatientRepo has no Delete; the guard must stop us before we get there.
	svc := NewPatientService(repo, fixedClock(now), &fakeDocs{count: 2})
	if err := svc.Delete(ctx, 1); !errors.Is(err, model.ErrConflict) {
		t.Errorf("got %v, want ErrConflict", err)
	}
}
