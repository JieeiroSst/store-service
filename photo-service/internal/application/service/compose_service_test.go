package service_test

import (
	"context"
	"errors"
	"image"
	"io"
	"testing"
	"time"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/application/service"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

// --- mocks for every outbound port -----------------------------------------

type mockJobRepository struct {
	SaveFunc     func(ctx context.Context, job *domain.CompositionJob) error
	FindByIDFunc func(ctx context.Context, id string) (*domain.CompositionJob, error)

	saved []*domain.CompositionJob
}

func (m *mockJobRepository) Save(ctx context.Context, job *domain.CompositionJob) error {
	m.saved = append(m.saved, job)
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, job)
	}
	return nil
}

func (m *mockJobRepository) FindByID(ctx context.Context, id string) (*domain.CompositionJob, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, domain.ErrJobNotFound
}

var _ ports.JobRepository = (*mockJobRepository)(nil)

type mockImageStorage struct {
	PutFunc          func(ctx context.Context, key string, r io.Reader, contentType string) error
	GetFunc          func(ctx context.Context, key string) (io.ReadCloser, error)
	PresignedURLFunc func(ctx context.Context, key string, ttl time.Duration) (string, error)
}

func (m *mockImageStorage) Put(ctx context.Context, key string, r io.Reader, contentType string) error {
	if m.PutFunc != nil {
		return m.PutFunc(ctx, key, r, contentType)
	}
	return nil
}

func (m *mockImageStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return io.NopCloser(nil), nil
}

func (m *mockImageStorage) PresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if m.PresignedURLFunc != nil {
		return m.PresignedURLFunc(ctx, key, ttl)
	}
	return "https://minio.local/" + key, nil
}

var _ ports.ImageStorage = (*mockImageStorage)(nil)

type mockCacheRepository struct {
	GetFunc func(ctx context.Context, key string) (string, error)
	SetFunc func(ctx context.Context, key, value string, ttl time.Duration) error
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", domain.ErrCacheMiss
}

func (m *mockCacheRepository) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, ttl)
	}
	return nil
}

var _ ports.CacheRepository = (*mockCacheRepository)(nil)

type mockImageComposer struct {
	ComposeFunc func(ctx context.Context, sources []image.Image, layout domain.LayoutConfig) ([]byte, error)
}

func (m *mockImageComposer) Compose(ctx context.Context, sources []image.Image, layout domain.LayoutConfig) ([]byte, error) {
	if m.ComposeFunc != nil {
		return m.ComposeFunc(ctx, sources, layout)
	}
	return []byte("fake-image-bytes"), nil
}

var _ ports.ImageComposer = (*mockImageComposer)(nil)

type mockImageDecoder struct {
	DecodeFunc func(ctx context.Context, data []byte) (image.Image, error)
}

func (m *mockImageDecoder) Decode(ctx context.Context, data []byte) (image.Image, error) {
	if m.DecodeFunc != nil {
		return m.DecodeFunc(ctx, data)
	}
	return image.NewRGBA(image.Rect(0, 0, 10, 10)), nil
}

var _ ports.ImageDecoder = (*mockImageDecoder)(nil)

type mockImageFetcher struct {
	FetchFunc func(ctx context.Context, url string) ([]byte, string, error)
}

func (m *mockImageFetcher) Fetch(ctx context.Context, url string) ([]byte, string, error) {
	if m.FetchFunc != nil {
		return m.FetchFunc(ctx, url)
	}
	return []byte("fetched-bytes"), "image/jpeg", nil
}

var _ ports.ImageFetcher = (*mockImageFetcher)(nil)

type noopLogger struct{}

func (noopLogger) Debug(string, ...any) {}
func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

var _ ports.Logger = noopLogger{}

// --- helpers -----------------------------------------------------------------

func newSources(n int) []domain.ImageSource {
	sources := make([]domain.ImageSource, n)
	for i := range sources {
		sources[i] = domain.ImageSource{
			Type:  domain.ImageSourceTypeUpload,
			Data:  []byte("data"),
			Order: i,
		}
	}
	return sources
}

func newService(jobs ports.JobRepository, storage ports.ImageStorage, cache ports.CacheRepository, composer ports.ImageComposer, decoder ports.ImageDecoder, fetcher ports.ImageFetcher) *service.ComposeService {
	return service.NewComposeService(jobs, storage, cache, composer, decoder, fetcher, noopLogger{})
}

// --- tests ---------------------------------------------------------------

func TestComposeImages_Success(t *testing.T) {
	jobs := &mockJobRepository{}

	svc := newService(jobs, &mockImageStorage{}, &mockCacheRepository{}, &mockImageComposer{}, &mockImageDecoder{}, &mockImageFetcher{})

	job, err := svc.ComposeImages(context.Background(), ports.ComposeImagesCommand{
		Sources: newSources(4),
		Layout:  domain.LayoutConfig{Type: domain.LayoutGrid, Cols: 2, Rows: 2, Width: 800, Height: 800},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.Status != domain.JobStatusCompleted {
		t.Fatalf("expected status completed, got %s (error: %s)", job.Status, job.ErrorMessage)
	}
	if job.ResultObjectKey == "" {
		t.Fatal("expected a result object key to be set")
	}
	if job.ResultURL == "" {
		t.Fatal("expected a result url to be set")
	}
	if job.Width != 800 || job.Height != 800 {
		t.Fatalf("expected job dimensions to match the layout canvas, got %dx%d", job.Width, job.Height)
	}
	if len(jobs.saved) < 1 {
		t.Fatalf("expected at least 1 job save call, got %d", len(jobs.saved))
	}
}

func TestComposeImages_NoSources(t *testing.T) {
	svc := newService(&mockJobRepository{}, &mockImageStorage{}, &mockCacheRepository{}, &mockImageComposer{}, &mockImageDecoder{}, &mockImageFetcher{})

	_, err := svc.ComposeImages(context.Background(), ports.ComposeImagesCommand{Sources: nil})
	if !errors.Is(err, domain.ErrNoSources) {
		t.Fatalf("expected ErrNoSources, got %v", err)
	}
}

func TestComposeImages_TooManySources(t *testing.T) {
	svc := newService(&mockJobRepository{}, &mockImageStorage{}, &mockCacheRepository{}, &mockImageComposer{}, &mockImageDecoder{}, &mockImageFetcher{})

	_, err := svc.ComposeImages(context.Background(), ports.ComposeImagesCommand{Sources: newSources(100)})
	if !errors.Is(err, domain.ErrTooManySources) {
		t.Fatalf("expected ErrTooManySources, got %v", err)
	}
}

func TestComposeImages_GridCellMismatchIsRejected(t *testing.T) {
	svc := newService(&mockJobRepository{}, &mockImageStorage{}, &mockCacheRepository{}, &mockImageComposer{}, &mockImageDecoder{}, &mockImageFetcher{})

	job, err := svc.ComposeImages(context.Background(), ports.ComposeImagesCommand{
		Sources: newSources(3), // 3 sources but a 2x2 = 4-cell grid
		Layout:  domain.LayoutConfig{Type: domain.LayoutGrid, Cols: 2, Rows: 2, Width: 800, Height: 800},
	})
	if err == nil {
		t.Fatalf("expected an error for source/cell count mismatch, got job %+v", job)
	}
	if !errors.Is(err, domain.ErrSourceCellMismatch) {
		t.Fatalf("expected ErrSourceCellMismatch, got %v", err)
	}
}

func TestComposeImages_ComposerFailureMarksJobFailed(t *testing.T) {
	jobs := &mockJobRepository{}
	composer := &mockImageComposer{
		ComposeFunc: func(ctx context.Context, sources []image.Image, layout domain.LayoutConfig) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	svc := newService(jobs, &mockImageStorage{}, &mockCacheRepository{}, composer, &mockImageDecoder{}, &mockImageFetcher{})

	job, err := svc.ComposeImages(context.Background(), ports.ComposeImagesCommand{
		Sources: newSources(2),
		Layout:  domain.LayoutConfig{Type: domain.LayoutGrid, Cols: 2, Rows: 1, Width: 800, Height: 400},
	})
	if err != nil {
		t.Fatalf("ComposeImages should not return an error when processing fails after job creation, got %v", err)
	}
	if job.Status != domain.JobStatusFailed {
		t.Fatalf("expected status failed, got %s", job.Status)
	}
	if job.ErrorMessage == "" {
		t.Fatal("expected error message to be recorded on the job")
	}
}

func TestComposeImages_IdempotentHitReturnsExistingJob(t *testing.T) {
	existing := &domain.CompositionJob{ID: "existing-job-id", Status: domain.JobStatusCompleted}

	jobs := &mockJobRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*domain.CompositionJob, error) {
			if id == existing.ID {
				return existing, nil
			}
			return nil, domain.ErrJobNotFound
		},
	}
	cache := &mockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			return existing.ID, nil
		},
	}

	svc := newService(jobs, &mockImageStorage{}, cache, &mockImageComposer{}, &mockImageDecoder{}, &mockImageFetcher{})

	job, err := svc.ComposeImages(context.Background(), ports.ComposeImagesCommand{
		Sources:        newSources(2),
		Layout:         domain.LayoutConfig{Type: domain.LayoutGrid, Cols: 2, Rows: 1, Width: 800, Height: 400},
		IdempotencyKey: "same-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != existing {
		t.Fatalf("expected the cached job to be returned, got a new job %+v", job)
	}
	if len(jobs.saved) != 0 {
		t.Fatalf("expected no new job to be saved on idempotent hit, got %d", len(jobs.saved))
	}
}

func TestComposeImages_ResolvesURLSourcesViaFetcherAndDecoder(t *testing.T) {
	var fetchedURL string
	var decodedBytes []byte
	fetcher := &mockImageFetcher{
		FetchFunc: func(ctx context.Context, url string) ([]byte, string, error) {
			fetchedURL = url
			return []byte("remote-bytes"), "image/png", nil
		},
	}
	decoder := &mockImageDecoder{
		DecodeFunc: func(ctx context.Context, data []byte) (image.Image, error) {
			decodedBytes = data
			return image.NewRGBA(image.Rect(0, 0, 10, 10)), nil
		},
	}

	svc := newService(&mockJobRepository{}, &mockImageStorage{}, &mockCacheRepository{}, &mockImageComposer{}, decoder, fetcher)

	sources := []domain.ImageSource{{Type: domain.ImageSourceTypeURL, URL: "https://example.com/a.jpg"}}
	job, err := svc.ComposeImages(context.Background(), ports.ComposeImagesCommand{
		Sources: sources,
		Layout:  domain.LayoutConfig{Type: domain.LayoutGrid, Cols: 1, Rows: 1, Width: 400, Height: 400},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetchedURL != "https://example.com/a.jpg" {
		t.Fatalf("expected fetcher to be called with the source url, got %q", fetchedURL)
	}
	if string(decodedBytes) != "remote-bytes" {
		t.Fatalf("expected decoder to receive the fetched bytes, got %q", decodedBytes)
	}
	if job.Status != domain.JobStatusCompleted {
		t.Fatalf("expected status completed, got %s (%s)", job.Status, job.ErrorMessage)
	}
}

func TestGetComposition_NotFound(t *testing.T) {
	jobs := &mockJobRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*domain.CompositionJob, error) {
			return nil, domain.ErrJobNotFound
		},
	}
	svc := newService(jobs, &mockImageStorage{}, &mockCacheRepository{}, &mockImageComposer{}, &mockImageDecoder{}, &mockImageFetcher{})

	_, err := svc.GetComposition(context.Background(), "missing")
	if !errors.Is(err, domain.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestGetComposition_RefreshesExpiredPresignedURL(t *testing.T) {
	completed := &domain.CompositionJob{
		ID:              "job-1",
		Status:          domain.JobStatusCompleted,
		ResultObjectKey: "compositions/job-1.jpg",
		ResultURL:       "https://minio.local/stale",
	}
	jobs := &mockJobRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*domain.CompositionJob, error) {
			return completed, nil
		},
	}
	cache := &mockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			return "", domain.ErrCacheMiss
		},
	}
	storage := &mockImageStorage{
		PresignedURLFunc: func(ctx context.Context, key string, ttl time.Duration) (string, error) {
			return "https://minio.local/fresh", nil
		},
	}

	svc := newService(jobs, storage, cache, &mockImageComposer{}, &mockImageDecoder{}, &mockImageFetcher{})

	job, err := svc.GetComposition(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ResultURL != "https://minio.local/fresh" {
		t.Fatalf("expected refreshed presigned url, got %q", job.ResultURL)
	}
}
