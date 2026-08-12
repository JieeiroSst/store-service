package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/JIeeiroSst/cdn-service/internal/ports"
	"github.com/google/uuid"
)

type fakeRepo struct {
	files map[uuid.UUID]model.File
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{files: make(map[uuid.UUID]model.File)}
}

func (r *fakeRepo) Create(ctx context.Context, f *model.File) error {
	f.CreatedAt = time.Now()
	f.UpdatedAt = f.CreatedAt
	r.files[f.ID] = *f
	return nil
}

func (r *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	f, ok := r.files[id]
	if !ok {
		return nil, nil
	}
	return &f, nil
}

func (r *fakeRepo) Update(ctx context.Context, f *model.File) error {
	f.UpdatedAt = time.Now()
	r.files[f.ID] = *f
	return nil
}

func (r *fakeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.files, id)
	return nil
}

func (r *fakeRepo) List(ctx context.Context, limit, offset int) ([]model.File, error) {
	out := make([]model.File, 0, len(r.files))
	for _, f := range r.files {
		out = append(out, f)
	}
	return out, nil
}

var errObjectNotFound = errors.New("object not found")

type fakeStorage struct {
	objects map[string]int64
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: make(map[string]int64)}
}

func (s *fakeStorage) PresignPutURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return "https://minio.local/put/" + objectKey, nil
}

func (s *fakeStorage) PresignGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return "https://minio.local/get/" + objectKey, nil
}

func (s *fakeStorage) StatObject(ctx context.Context, objectKey string) (int64, error) {
	size, ok := s.objects[objectKey]
	if !ok {
		return 0, errObjectNotFound
	}
	return size, nil
}

func (s *fakeStorage) DeleteObject(ctx context.Context, objectKey string) error {
	delete(s.objects, objectKey)
	return nil
}

func newTestService() (*FileService, *fakeRepo, *fakeStorage) {
	repo := newFakeRepo()
	storage := newFakeStorage()
	svc := NewFileService(repo, storage, FileServiceOptions{
		MaxUploadSizeBytes: 1024,
		PresignPutExpiry:   time.Minute,
		PresignGetExpiry:   time.Minute,
		EdgeCacheBaseURL:   "https://cdn.internal.company.vn",
	})
	return svc, repo, storage
}

func TestCreatePresignedUpload(t *testing.T) {
	svc, _, _ := newTestService()

	pu, err := svc.CreatePresignedUpload(context.Background(), ports.CreatePresignInput{
		FileName:    "logo.png",
		ContentType: "image/png",
		SizeBytes:   512,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pu.File.Status != model.FileStatusPending {
		t.Errorf("expected status pending, got %s", pu.File.Status)
	}
	if pu.UploadURL == "" {
		t.Error("expected non-empty upload url")
	}
}

func TestCreatePresignedUpload_TooLarge(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.CreatePresignedUpload(context.Background(), ports.CreatePresignInput{
		FileName:    "big.bin",
		ContentType: "application/octet-stream",
		SizeBytes:   2048,
	})
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestConfirmUploadAndDownloadURL(t *testing.T) {
	svc, _, storage := newTestService()
	ctx := context.Background()

	pu, err := svc.CreatePresignedUpload(ctx, ports.CreatePresignInput{
		FileName:    "logo.png",
		ContentType: "image/png",
		SizeBytes:   512,
	})
	if err != nil {
		t.Fatalf("presign: %v", err)
	}

	if _, err := svc.GetDownloadURL(ctx, pu.File.ID, false); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus before confirm, got %v", err)
	}

	storage.objects[pu.File.ObjectKey] = 512

	confirmed, err := svc.ConfirmUpload(ctx, pu.File.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Status != model.FileStatusUploaded {
		t.Fatalf("expected uploaded status, got %s", confirmed.Status)
	}

	edgeURL, err := svc.GetDownloadURL(ctx, pu.File.ID, false)
	if err != nil {
		t.Fatalf("download url: %v", err)
	}
	want := "https://cdn.internal.company.vn/" + pu.File.ObjectKey
	if edgeURL != want {
		t.Errorf("expected edge url %q, got %q", want, edgeURL)
	}

	directURL, err := svc.GetDownloadURL(ctx, pu.File.ID, true)
	if err != nil {
		t.Fatalf("direct download url: %v", err)
	}
	if directURL == edgeURL {
		t.Errorf("expected direct url to differ from edge url")
	}
}

func TestDeleteFile(t *testing.T) {
	svc, repo, storage := newTestService()
	ctx := context.Background()

	pu, err := svc.CreatePresignedUpload(ctx, ports.CreatePresignInput{
		FileName:    "logo.png",
		ContentType: "image/png",
		SizeBytes:   512,
	})
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	storage.objects[pu.File.ObjectKey] = 512

	if err := svc.DeleteFile(ctx, pu.File.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := repo.files[pu.File.ID]; ok {
		t.Error("expected file metadata to be removed")
	}
	if _, ok := storage.objects[pu.File.ObjectKey]; ok {
		t.Error("expected object to be removed from storage")
	}
}
