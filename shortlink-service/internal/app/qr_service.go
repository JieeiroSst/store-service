package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

// QRService mirrors src/routes/qr.ts.
type QRService struct {
	links     ports.LinkRepository
	cache     ports.Cache
	generator ports.QRCodeGenerator
}

func NewQRService(links ports.LinkRepository, cache ports.Cache, generator ports.QRCodeGenerator) *QRService {
	return &QRService{links, cache, generator}
}

type QRInput struct {
	LinkID          string
	Format          string // "png" | "svg"
	Size            int
	Color           string
	BGColor         string
	ShortLinkDomain string // SHORTLINK_DOMAIN env, or request-derived origin
}

type QRResult struct {
	ContentType string
	PNG         []byte
	SVG         string
	ShortCode   string
}

var ErrInvalidQRFormat = errors.New("invalid format")

func (s *QRService) Generate(ctx context.Context, in QRInput) (*QRResult, error) {
	if in.Format != "png" && in.Format != "svg" {
		return nil, ErrInvalidQRFormat
	}
	size := in.Size
	if size < 128 {
		size = 128
	}
	if size > 2048 {
		size = 2048
	}
	color := in.Color
	if color == "" {
		color = "#000000"
	}
	bgcolor := in.BGColor
	if bgcolor == "" {
		bgcolor = "#ffffff"
	}

	cacheKey := fmt.Sprintf("qr:%s:%s:%s:%s:%s", in.LinkID, in.Format, strconv.Itoa(size), color, bgcolor)

	if s.cache.Enabled() {
		if cached, ok := s.cache.Get(ctx, cacheKey); ok {
			if in.Format == "png" {
				return &QRResult{ContentType: "image/png", PNG: []byte(cached)}, nil
			}
			return &QRResult{ContentType: "image/svg+xml", SVG: cached}, nil
		}
	}

	link, err := s.links.GetByID(ctx, in.LinkID, ports.LinkFilter{})
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}
	if !link.IsActive {
		return nil, ErrLinkNotFound
	}

	shortURL := link.OriginalURL
	if link.ShortCode != "" {
		shortURL = in.ShortLinkDomain + "/" + link.ShortCode
	}

	if in.Format == "png" {
		png, err := s.generator.PNG(shortURL, size, color, bgcolor)
		if err != nil {
			return nil, err
		}
		if s.cache.Enabled() {
			s.cache.Set(ctx, cacheKey, string(png), 24*time.Hour)
		}
		return &QRResult{ContentType: "image/png", PNG: png, ShortCode: link.ShortCode}, nil
	}

	svg, err := s.generator.SVG(shortURL, size, color, bgcolor)
	if err != nil {
		return nil, err
	}
	if s.cache.Enabled() {
		s.cache.Set(ctx, cacheKey, svg, 24*time.Hour)
	}
	return &QRResult{ContentType: "image/svg+xml", SVG: svg, ShortCode: link.ShortCode}, nil
}
