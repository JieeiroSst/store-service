package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

const (
	maxSources        = 64
	defaultPresignTTL = time.Hour
	idempotencyTTL    = 24 * time.Hour

	presignedURLCachePrefix = "presigned:"
	idempotencyCachePrefix  = "idempotency:"
)

type ComposeService struct {
	jobs     ports.JobRepository
	storage  ports.ImageStorage
	cache    ports.CacheRepository
	composer ports.ImageComposer
	decoder  ports.ImageDecoder
	fetcher  ports.ImageFetcher
	log      ports.Logger
}

func NewComposeService(
	jobs ports.JobRepository,
	storage ports.ImageStorage,
	cache ports.CacheRepository,
	composer ports.ImageComposer,
	decoder ports.ImageDecoder,
	fetcher ports.ImageFetcher,
	log ports.Logger,
) *ComposeService {
	return &ComposeService{
		jobs:     jobs,
		storage:  storage,
		cache:    cache,
		composer: composer,
		decoder:  decoder,
		fetcher:  fetcher,
		log:      log,
	}
}

var _ ports.ComposeImageUseCase = (*ComposeService)(nil)

func (s *ComposeService) ComposeImages(ctx context.Context, cmd ports.ComposeImagesCommand) (*domain.CompositionJob, error) {
	if len(cmd.Sources) == 0 {
		return nil, domain.ErrNoSources
	}
	if len(cmd.Sources) > maxSources {
		return nil, domain.ErrTooManySources
	}

	if cmd.IdempotencyKey != "" {
		if existing, err := s.findByIdempotencyKey(ctx, cmd.IdempotencyKey); err != nil {
			return nil, err
		} else if existing != nil {
			s.log.Info("composition idempotent hit", "idempotency_key", cmd.IdempotencyKey, "job_id", existing.ID)
			return existing, nil
		}
	}

	layout := cmd.Layout
	layout.Normalize(len(cmd.Sources))
	if err := layout.Validate(len(cmd.Sources)); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	job := &domain.CompositionJob{
		ID:             uuid.NewString(),
		Status:         domain.JobStatusPending,
		Layout:         layout,
		Sources:        cmd.Sources,
		IdempotencyKey: cmd.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.jobs.Save(ctx, job); err != nil {
		return nil, fmt.Errorf("save job: %w", err)
	}

	if cmd.IdempotencyKey != "" {
		if err := s.cache.Set(ctx, idempotencyCachePrefix+cmd.IdempotencyKey, job.ID, idempotencyTTL); err != nil {
			s.log.Warn("failed to cache idempotency key", "error", err, "job_id", job.ID)
		}
	}

	s.process(ctx, job)

	return job, nil
}


func (s *ComposeService) process(ctx context.Context, job *domain.CompositionJob) {
	job.MarkProcessing()
	if err := s.jobs.Save(ctx, job); err != nil {
		s.log.Error("failed to mark job processing", "error", err, "job_id", job.ID)
	}

	images, err := s.resolveImages(ctx, job.Sources)
	if err != nil {
		s.fail(ctx, job, err)
		return
	}

	data, err := s.composer.Compose(ctx, images, job.Layout)
	if err != nil {
		s.fail(ctx, job, fmt.Errorf("%w: %v", domain.ErrComposeFailed, err))
		return
	}

	objectKey := fmt.Sprintf("compositions/%s.%s", job.ID, extensionFor(job.Layout.Format))
	contentType := contentTypeFor(job.Layout.Format)
	if err := s.storage.Put(ctx, objectKey, bytes.NewReader(data), contentType); err != nil {
		s.fail(ctx, job, fmt.Errorf("upload result: %w", err))
		return
	}

	presigned, err := s.storage.PresignedURL(ctx, objectKey, defaultPresignTTL)
	if err != nil {
		s.fail(ctx, job, fmt.Errorf("presign result url: %w", err))
		return
	}
	if err := s.cache.Set(ctx, presignedURLCachePrefix+objectKey, presigned, defaultPresignTTL); err != nil {
		s.log.Warn("failed to cache presigned url", "error", err, "object_key", objectKey)
	}

	job.MarkCompleted(objectKey, presigned, job.Layout.Width, job.Layout.Height, job.Layout.Format, int64(len(data)))
	if err := s.jobs.Save(ctx, job); err != nil {
		s.log.Error("failed to persist completed job", "error", err, "job_id", job.ID)
	}
}

func (s *ComposeService) fail(ctx context.Context, job *domain.CompositionJob, cause error) {
	s.log.Error("composition failed", "error", cause, "job_id", job.ID)
	job.MarkFailed(cause)
	if err := s.jobs.Save(ctx, job); err != nil {
		s.log.Error("failed to persist failed job", "error", err, "job_id", job.ID)
	}
}

func (s *ComposeService) resolveImages(ctx context.Context, sources []domain.ImageSource) ([]image.Image, error) {
	ordered := make([]domain.ImageSource, len(sources))
	copy(ordered, sources)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Order < ordered[j].Order })

	images := make([]image.Image, len(ordered))
	for i, src := range ordered {
		data, err := s.resolveBytes(ctx, src)
		if err != nil {
			return nil, err
		}
		img, err := s.decoder.Decode(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("%w: source %d: %v", domain.ErrInvalidImageData, i, err)
		}
		images[i] = img
	}
	return images, nil
}

func (s *ComposeService) resolveBytes(ctx context.Context, src domain.ImageSource) ([]byte, error) {
	switch src.Type {
	case domain.ImageSourceTypeURL:
		data, _, err := s.fetcher.Fetch(ctx, src.URL)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", domain.ErrFetchSourceFailed, src.URL, err)
		}
		return data, nil
	case domain.ImageSourceTypeMinIO:
		rc, err := s.storage.Get(ctx, src.ObjectKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", domain.ErrFetchSourceFailed, src.ObjectKey, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", domain.ErrFetchSourceFailed, src.ObjectKey, err)
		}
		return data, nil
	default:
		if len(src.Data) == 0 {
			return nil, domain.ErrInvalidImageData
		}
		return src.Data, nil
	}
}

func (s *ComposeService) findByIdempotencyKey(ctx context.Context, key string) (*domain.CompositionJob, error) {
	jobID, err := s.cache.Get(ctx, idempotencyCachePrefix+key)
	if err != nil {
		if !errors.Is(err, domain.ErrCacheMiss) {
			s.log.Warn("idempotency cache lookup failed", "error", err, "idempotency_key", key)
		}
		return nil, nil
	}

	job, err := s.jobs.FindByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup idempotent job: %w", err)
	}
	return job, nil
}

func (s *ComposeService) GetComposition(ctx context.Context, id string) (*domain.CompositionJob, error) {
	job, err := s.jobs.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if job.Status != domain.JobStatusCompleted || job.ResultObjectKey == "" {
		return job, nil
	}

	if cached, err := s.cache.Get(ctx, presignedURLCachePrefix+job.ResultObjectKey); err == nil {
		job.ResultURL = cached
		return job, nil
	}

	presigned, err := s.storage.PresignedURL(ctx, job.ResultObjectKey, defaultPresignTTL)
	if err != nil {
		s.log.Warn("failed to refresh presigned url", "error", err, "object_key", job.ResultObjectKey)
		return job, nil
	}
	job.ResultURL = presigned
	if err := s.cache.Set(ctx, presignedURLCachePrefix+job.ResultObjectKey, presigned, defaultPresignTTL); err != nil {
		s.log.Warn("failed to cache refreshed presigned url", "error", err, "object_key", job.ResultObjectKey)
	}

	return job, nil
}

func extensionFor(f domain.OutputFormat) string {
	switch f {
	case domain.OutputFormatPNG:
		return "png"
	case domain.OutputFormatWebP:
		return "webp"
	default:
		return "jpg"
	}
}

func contentTypeFor(f domain.OutputFormat) string {
	switch f {
	case domain.OutputFormatPNG:
		return "image/png"
	case domain.OutputFormatWebP:
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
