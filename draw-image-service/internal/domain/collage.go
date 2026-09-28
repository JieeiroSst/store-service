package domain

import (
	"errors"
	"io"
	"time"
)

type Layout string

const (
	LayoutRow    Layout = "row"
	LayoutGrid   Layout = "grid"
	LayoutCenter Layout = "center"
)

var (
	ErrNoImages      = errors.New("no images provided")
	ErrTooManyImages = errors.New("too many images")
	ErrInvalidLayout = errors.New("invalid layout")
	ErrInvalidImage  = errors.New("invalid image")
	ErrImageTooLarge = errors.New("image dimensions too large")
	ErrInvalidID     = errors.New("invalid collage id")
	ErrNotFound      = errors.New("collage not found")
)

type Source struct {
	Name    string
	Content io.Reader
}

type CollageRequest struct {
	Sources []Source
	Layout  Layout
	Columns int
}

type Collage struct {
	ID          string
	Bucket      string
	ObjectKey   string
	ContentType string
	Size        int64
	Width       int
	Height      int
	ImageCount  int
	Layout      Layout
	Columns     int
	CreatedAt   time.Time
}

func ParseLayout(raw string) (Layout, error) {
	switch Layout(raw) {
	case "", LayoutGrid:
		return LayoutGrid, nil
	case LayoutRow:
		return LayoutRow, nil
	case LayoutCenter:
		return LayoutCenter, nil
	}
	return "", ErrInvalidLayout
}
