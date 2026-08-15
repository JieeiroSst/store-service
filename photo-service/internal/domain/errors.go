package domain

import "errors"

var (
	ErrNoSources          = errors.New("domain: at least one image source is required")
	ErrTooManySources     = errors.New("domain: too many image sources for a single composition")
	ErrInvalidLayout      = errors.New("domain: invalid layout configuration")
	ErrInvalidLayoutType  = errors.New("domain: invalid layout type")
	ErrInvalidCellFit     = errors.New("domain: invalid cell fit mode")
	ErrInvalidCellSpec    = errors.New("domain: invalid cell specification")
	ErrSourceCellMismatch = errors.New("domain: number of image sources does not match number of layout cells")
	ErrUnsupportedFormat  = errors.New("domain: unsupported output format")
	ErrJobNotFound        = errors.New("domain: composition job not found")
	ErrComposeFailed      = errors.New("domain: failed to compose image")
	ErrInvalidImageData   = errors.New("domain: invalid or corrupt image data")
	ErrFetchSourceFailed  = errors.New("domain: failed to fetch image source from url")
	ErrCacheMiss          = errors.New("domain: cache key not found")
)
