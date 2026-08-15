package domain

import (
	"fmt"
	"time"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type ImageSourceType string

const (
	ImageSourceTypeUpload ImageSourceType = "upload"
	ImageSourceTypeURL    ImageSourceType = "url"
	ImageSourceTypeMinIO  ImageSourceType = "minio"
)

type OutputFormat string

const (
	OutputFormatJPEG OutputFormat = "jpeg"
	OutputFormatPNG  OutputFormat = "png"
	OutputFormatWebP OutputFormat = "webp"
)

type LayoutType string

const (
	LayoutGrid     LayoutType = "grid"
	LayoutMosaic   LayoutType = "mosaic"
	LayoutFreeform LayoutType = "freeform"
)

type CellFit string

const (
	FitCover   CellFit = "cover"
	FitContain CellFit = "contain"
	FitStretch CellFit = "stretch"
)

type CellSpec struct {
	X, Y          int
	Width, Height int
	ImageIndex    int
}

type LayoutConfig struct {
	Type       LayoutType
	Rows, Cols int
	Width      int
	Height     int
	Spacing    int
	Padding    int
	Background string
	CellFit    CellFit
	Cells      []CellSpec
	Format     OutputFormat
	Quality    int
}

const (
	defaultCanvasSize = 1200
	defaultQuality    = 90
)

func (l *LayoutConfig) Normalize(sourceCount int) {
	if l.Type == "" {
		l.Type = LayoutGrid
	}
	if l.Width <= 0 {
		l.Width = defaultCanvasSize
	}
	if l.Height <= 0 {
		l.Height = defaultCanvasSize
	}
	if l.Background == "" {
		l.Background = "#FFFFFF"
	}
	if l.CellFit == "" {
		l.CellFit = FitCover
	}
	if l.Format == "" {
		l.Format = OutputFormatJPEG
	}
	if l.Quality <= 0 {
		l.Quality = defaultQuality
	}

	if l.Type == LayoutGrid {
		switch {
		case l.Rows <= 0 && l.Cols <= 0:
			cols := ceilSqrt(sourceCount)
			l.Cols = cols
			l.Rows = (sourceCount + cols - 1) / cols
		case l.Cols <= 0:
			l.Cols = (sourceCount + l.Rows - 1) / l.Rows
		case l.Rows <= 0:
			l.Rows = (sourceCount + l.Cols - 1) / l.Cols
		}
	}
}

func ceilSqrt(n int) int {
	if n <= 0 {
		return 1
	}
	c := 1
	for c*c < n {
		c++
	}
	return c
}

func (l LayoutConfig) Validate(sourceCount int) error {
	if sourceCount <= 0 {
		return ErrNoSources
	}
	if l.Width <= 0 || l.Height <= 0 {
		return fmt.Errorf("%w: canvas width/height must be positive", ErrInvalidLayout)
	}
	switch l.CellFit {
	case FitCover, FitContain, FitStretch:
	default:
		return ErrInvalidCellFit
	}
	switch l.Format {
	case OutputFormatJPEG, OutputFormatPNG, OutputFormatWebP:
	default:
		return ErrUnsupportedFormat
	}
	if l.Quality < 1 || l.Quality > 100 {
		return fmt.Errorf("%w: quality must be between 1 and 100", ErrInvalidLayout)
	}

	_, err := l.ResolveCells(sourceCount)
	return err
}

func (l LayoutConfig) ResolveCells(sourceCount int) ([]CellSpec, error) {
	switch l.Type {
	case LayoutGrid, "":
		return l.gridCells(sourceCount)
	case LayoutMosaic, LayoutFreeform:
		return l.explicitCells(sourceCount)
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidLayoutType, l.Type)
	}
}

func (l LayoutConfig) gridCells(sourceCount int) ([]CellSpec, error) {
	if l.Rows <= 0 || l.Cols <= 0 {
		return nil, fmt.Errorf("%w: grid layout requires positive rows and cols", ErrInvalidLayout)
	}
	if l.Rows*l.Cols != sourceCount {
		return nil, fmt.Errorf("%w: grid has %d cells (%dx%d), got %d sources", ErrSourceCellMismatch, l.Rows*l.Cols, l.Cols, l.Rows, sourceCount)
	}

	cellAreaW := l.Width - 2*l.Padding - (l.Cols-1)*l.Spacing
	cellAreaH := l.Height - 2*l.Padding - (l.Rows-1)*l.Spacing
	if cellAreaW <= 0 || cellAreaH <= 0 {
		return nil, fmt.Errorf("%w: %dx%d canvas is too small for a %dx%d grid with padding=%d spacing=%d", ErrInvalidLayout, l.Width, l.Height, l.Cols, l.Rows, l.Padding, l.Spacing)
	}
	cellW := cellAreaW / l.Cols
	cellH := cellAreaH / l.Rows
	if cellW <= 0 || cellH <= 0 {
		return nil, fmt.Errorf("%w: computed grid cell size is zero", ErrInvalidLayout)
	}

	cells := make([]CellSpec, sourceCount)
	for i := 0; i < sourceCount; i++ {
		col := i % l.Cols
		row := i / l.Cols
		cells[i] = CellSpec{
			X:          l.Padding + col*(cellW+l.Spacing),
			Y:          l.Padding + row*(cellH+l.Spacing),
			Width:      cellW,
			Height:     cellH,
			ImageIndex: i,
		}
	}
	return cells, nil
}

func (l LayoutConfig) explicitCells(sourceCount int) ([]CellSpec, error) {
	if len(l.Cells) == 0 {
		return nil, fmt.Errorf("%w: %s layout requires at least one cell", ErrInvalidLayout, l.Type)
	}
	if len(l.Cells) != sourceCount {
		return nil, fmt.Errorf("%w: layout defines %d cells, got %d sources", ErrSourceCellMismatch, len(l.Cells), sourceCount)
	}

	seenIndex := make(map[int]bool, len(l.Cells))
	for _, cell := range l.Cells {
		if cell.Width <= 0 || cell.Height <= 0 {
			return nil, fmt.Errorf("%w: cell has non-positive width/height", ErrInvalidCellSpec)
		}
		if cell.X < 0 || cell.Y < 0 || cell.X+cell.Width > l.Width || cell.Y+cell.Height > l.Height {
			return nil, fmt.Errorf("%w: cell (x=%d,y=%d,w=%d,h=%d) falls outside the %dx%d canvas", ErrInvalidCellSpec, cell.X, cell.Y, cell.Width, cell.Height, l.Width, l.Height)
		}
		if cell.ImageIndex < 0 || cell.ImageIndex >= sourceCount {
			return nil, fmt.Errorf("%w: cell references image index %d, have %d sources", ErrInvalidCellSpec, cell.ImageIndex, sourceCount)
		}
		if seenIndex[cell.ImageIndex] {
			return nil, fmt.Errorf("%w: image index %d is referenced by more than one cell", ErrInvalidCellSpec, cell.ImageIndex)
		}
		seenIndex[cell.ImageIndex] = true
	}
	return l.Cells, nil
}

type ImageSource struct {
	Type        ImageSourceType
	URL         string
	ObjectKey   string
	Data        []byte
	ContentType string
	Order       int
}

type CompositionJob struct {
	ID              string
	Status          JobStatus
	Layout          LayoutConfig
	Sources         []ImageSource
	IdempotencyKey  string
	ResultObjectKey string
	ResultURL       string
	Width           int
	Height          int
	Format          OutputFormat
	SizeBytes       int64
	ErrorMessage    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (j *CompositionJob) MarkProcessing() {
	j.Status = JobStatusProcessing
	j.UpdatedAt = time.Now().UTC()
}

func (j *CompositionJob) MarkFailed(err error) {
	j.Status = JobStatusFailed
	j.ErrorMessage = err.Error()
	j.UpdatedAt = time.Now().UTC()
}

func (j *CompositionJob) MarkCompleted(objectKey, url string, width, height int, format OutputFormat, size int64) {
	j.Status = JobStatusCompleted
	j.ResultObjectKey = objectKey
	j.ResultURL = url
	j.Width = width
	j.Height = height
	j.Format = format
	j.SizeBytes = size
	j.UpdatedAt = time.Now().UTC()
}
