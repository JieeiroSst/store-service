package application

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"log"
	"time"

	"github.com/JIeeiroSst/draw-image-service/config"
	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"github.com/JIeeiroSst/draw-image-service/internal/port"
	"github.com/google/uuid"
)

const contentTypeJPEG = "image/jpeg"

type collageService struct {
	processor port.ImageProcessor
	storage   port.ObjectStorage
	repo      port.CollageRepository
	cfg       config.CollageConfig
	now       func() time.Time
}

func NewCollageService(
	processor port.ImageProcessor,
	storage port.ObjectStorage,
	repo port.CollageRepository,
	cfg *config.Config,
) port.CollageService {
	return &collageService{
		processor: processor,
		storage:   storage,
		repo:      repo,
		cfg:       cfg.Collage,
		now:       time.Now,
	}
}

func (s *collageService) Create(ctx context.Context, req domain.CollageRequest) (*domain.Collage, error) {
	count := len(req.Sources)
	if count == 0 {
		return nil, domain.ErrNoImages
	}
	if count > s.cfg.MaxImages {
		return nil, fmt.Errorf("%w: %d > %d", domain.ErrTooManyImages, count, s.cfg.MaxImages)
	}

	layout, columns, err := s.columns(req, count)
	if err != nil {
		return nil, err
	}

	images := make([]image.Image, 0, count)
	for _, src := range req.Sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		img, err := s.processor.Decode(src.Content)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", src.Name, err)
		}
		if layout == domain.LayoutCenter {
			img = s.processor.Shrink(img, s.cfg.CenterWidth, s.cfg.CenterHeight)
		} else {
			img = s.processor.Thumbnail(img, s.cfg.CellWidth, s.cfg.CellHeight)
		}
		images = append(images, img)
	}

	var dst image.Image
	if layout == domain.LayoutCenter {
		dst = s.processor.ComposeCenter(images[0], images[1:], s.cfg.CenterWidth, s.cfg.CenterHeight)
	} else {
		dst = s.processor.Compose(images, s.cfg.CellWidth, s.cfg.CellHeight, columns)
	}
	var buf bytes.Buffer
	if err := s.processor.EncodeJPEG(&buf, dst, s.cfg.JPEGQuality); err != nil {
		return nil, fmt.Errorf("encode collage: %w", err)
	}

	id := uuid.NewString()
	bounds := dst.Bounds()
	c := &domain.Collage{
		ID:          id,
		Bucket:      s.storage.Bucket(),
		ObjectKey:   objectKey(id),
		ContentType: contentTypeJPEG,
		Size:        int64(buf.Len()),
		Width:       bounds.Dx(),
		Height:      bounds.Dy(),
		ImageCount:  count,
		Layout:      layout,
		Columns:     columns,
		CreatedAt:   s.now().UTC(),
	}

	if err := s.storage.Put(ctx, c.ObjectKey, &buf, c.Size, c.ContentType); err != nil {
		return nil, fmt.Errorf("store collage: %w", err)
	}
	if err := s.repo.Save(ctx, c); err != nil {
		if delErr := s.storage.Delete(context.WithoutCancel(ctx), c.ObjectKey); delErr != nil {
			log.Printf("cleanup object %s: %v", c.ObjectKey, delErr)
		}
		return nil, fmt.Errorf("save collage: %w", err)
	}
	return c, nil
}

func (s *collageService) Get(ctx context.Context, id string) (*domain.Collage, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.ErrInvalidID
	}
	return s.repo.FindByID(ctx, id)
}

func (s *collageService) Open(ctx context.Context, id string) (*domain.Collage, io.ReadCloser, error) {
	c, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	rc, err := s.storage.Get(ctx, c.ObjectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("load collage %s: %w", id, err)
	}
	return c, rc, nil
}

func (s *collageService) columns(req domain.CollageRequest, count int) (domain.Layout, int, error) {
	switch req.Layout {
	case domain.LayoutRow:
		return domain.LayoutRow, count, nil
	case domain.LayoutCenter:
		return domain.LayoutCenter, 0, nil
	case domain.LayoutGrid, "":
		columns := req.Columns
		if columns <= 0 {
			columns = s.cfg.Columns
		}
		return domain.LayoutGrid, min(columns, count), nil
	}
	return "", 0, domain.ErrInvalidLayout
}

func objectKey(id string) string {
	return "collages/" + id + ".jpg"
}
