package http

import (
	"time"

	"github.com/JIeeiroSst/photo-service/internal/domain"
)

type cellDTO struct {
	X          int `json:"x"`
	Y          int `json:"y"`
	Width      int `json:"width"`
	Height     int `json:"height"`
	ImageIndex int `json:"image_index"`
}

func (d cellDTO) toDomain() domain.CellSpec {
	return domain.CellSpec{X: d.X, Y: d.Y, Width: d.Width, Height: d.Height, ImageIndex: d.ImageIndex}
}

type layoutDTO struct {
	Type       string    `json:"type"`
	Rows       int       `json:"rows"`
	Cols       int       `json:"cols"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Spacing    int       `json:"spacing"`
	Padding    int       `json:"padding"`
	Background string    `json:"background"`
	CellFit    string    `json:"cell_fit"`
	Cells      []cellDTO `json:"cells,omitempty"`
	Format     string    `json:"format"`
	Quality    int       `json:"quality"`
}

func (d layoutDTO) toDomain() domain.LayoutConfig {
	cells := make([]domain.CellSpec, len(d.Cells))
	for i, c := range d.Cells {
		cells[i] = c.toDomain()
	}
	return domain.LayoutConfig{
		Type:       domain.LayoutType(d.Type),
		Rows:       d.Rows,
		Cols:       d.Cols,
		Width:      d.Width,
		Height:     d.Height,
		Spacing:    d.Spacing,
		Padding:    d.Padding,
		Background: d.Background,
		CellFit:    domain.CellFit(d.CellFit),
		Cells:      cells,
		Format:     domain.OutputFormat(d.Format),
		Quality:    d.Quality,
	}
}

func newLayoutDTO(l domain.LayoutConfig) layoutDTO {
	cells := make([]cellDTO, len(l.Cells))
	for i, c := range l.Cells {
		cells[i] = cellDTO{X: c.X, Y: c.Y, Width: c.Width, Height: c.Height, ImageIndex: c.ImageIndex}
	}
	return layoutDTO{
		Type:       string(l.Type),
		Rows:       l.Rows,
		Cols:       l.Cols,
		Width:      l.Width,
		Height:     l.Height,
		Spacing:    l.Spacing,
		Padding:    l.Padding,
		Background: l.Background,
		CellFit:    string(l.CellFit),
		Cells:      cells,
		Format:     string(l.Format),
		Quality:    l.Quality,
	}
}

type sourceDTO struct {
	Type      string `json:"type"`
	URL       string `json:"url,omitempty"`
	ObjectKey string `json:"object_key,omitempty"`
	Order     int    `json:"order"`
}

type compositionResponse struct {
	ID           string      `json:"id"`
	Status       string      `json:"status"`
	Layout       layoutDTO   `json:"layout"`
	Sources      []sourceDTO `json:"sources"`
	ObjectKey    string      `json:"object_key,omitempty"`
	URL          string      `json:"url,omitempty"`
	Width        int         `json:"width,omitempty"`
	Height       int         `json:"height,omitempty"`
	Format       string      `json:"format,omitempty"`
	SizeBytes    int64       `json:"size_bytes,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func newCompositionResponse(job *domain.CompositionJob) compositionResponse {
	sources := make([]sourceDTO, len(job.Sources))
	for i, s := range job.Sources {
		sources[i] = sourceDTO{Type: string(s.Type), URL: s.URL, ObjectKey: s.ObjectKey, Order: s.Order}
	}
	return compositionResponse{
		ID:           job.ID,
		Status:       string(job.Status),
		Layout:       newLayoutDTO(job.Layout),
		Sources:      sources,
		ObjectKey:    job.ResultObjectKey,
		URL:          job.ResultURL,
		Width:        job.Width,
		Height:       job.Height,
		Format:       string(job.Format),
		SizeBytes:    job.SizeBytes,
		ErrorMessage: job.ErrorMessage,
		CreatedAt:    job.CreatedAt,
		UpdatedAt:    job.UpdatedAt,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}
