package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

type JobRepository struct {
	pool *pgxpool.Pool
}

func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{pool: pool}
}

var _ ports.JobRepository = (*JobRepository)(nil)

type cellRecord struct {
	X          int `json:"x"`
	Y          int `json:"y"`
	Width      int `json:"width"`
	Height     int `json:"height"`
	ImageIndex int `json:"image_index"`
}

type layoutRecord struct {
	Type       string       `json:"type"`
	Rows       int          `json:"rows"`
	Cols       int          `json:"cols"`
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	Spacing    int          `json:"spacing"`
	Padding    int          `json:"padding"`
	Background string       `json:"background"`
	CellFit    string       `json:"cell_fit"`
	Cells      []cellRecord `json:"cells,omitempty"`
	Format     string       `json:"format"`
	Quality    int          `json:"quality"`
}

type sourceRecord struct {
	Type        string `json:"type"`
	URL         string `json:"url,omitempty"`
	ObjectKey   string `json:"object_key,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Order       int    `json:"order"`
}

func toLayoutRecord(l domain.LayoutConfig) layoutRecord {
	cells := make([]cellRecord, len(l.Cells))
	for i, c := range l.Cells {
		cells[i] = cellRecord{X: c.X, Y: c.Y, Width: c.Width, Height: c.Height, ImageIndex: c.ImageIndex}
	}
	return layoutRecord{
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

func (r layoutRecord) toDomain() domain.LayoutConfig {
	cells := make([]domain.CellSpec, len(r.Cells))
	for i, c := range r.Cells {
		cells[i] = domain.CellSpec{X: c.X, Y: c.Y, Width: c.Width, Height: c.Height, ImageIndex: c.ImageIndex}
	}
	return domain.LayoutConfig{
		Type:       domain.LayoutType(r.Type),
		Rows:       r.Rows,
		Cols:       r.Cols,
		Width:      r.Width,
		Height:     r.Height,
		Spacing:    r.Spacing,
		Padding:    r.Padding,
		Background: r.Background,
		CellFit:    domain.CellFit(r.CellFit),
		Cells:      cells,
		Format:     domain.OutputFormat(r.Format),
		Quality:    r.Quality,
	}
}

type sourceRecords []sourceRecord

func toSourceRecords(sources []domain.ImageSource) sourceRecords {
	out := make(sourceRecords, len(sources))
	for i, s := range sources {
		out[i] = sourceRecord{
			Type:        string(s.Type),
			URL:         s.URL,
			ObjectKey:   s.ObjectKey,
			ContentType: s.ContentType,
			Order:       s.Order,
		}
	}
	return out
}

func (recs sourceRecords) toDomain() []domain.ImageSource {
	out := make([]domain.ImageSource, len(recs))
	for i, r := range recs {
		out[i] = domain.ImageSource{
			Type:        domain.ImageSourceType(r.Type),
			URL:         r.URL,
			ObjectKey:   r.ObjectKey,
			ContentType: r.ContentType,
			Order:       r.Order,
		}
	}
	return out
}

const jobColumns = `id, status, layout, sources, idempotency_key, result_object_key, result_url, width, height, format, size_bytes, error_message, created_at, updated_at`

func (r *JobRepository) Save(ctx context.Context, job *domain.CompositionJob) error {
	layoutJSON, err := json.Marshal(toLayoutRecord(job.Layout))
	if err != nil {
		return fmt.Errorf("marshal layout: %w", err)
	}
	sourcesJSON, err := json.Marshal(toSourceRecords(job.Sources))
	if err != nil {
		return fmt.Errorf("marshal sources: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO composition_jobs (
			id, status, layout, sources, idempotency_key, result_object_key, result_url,
			width, height, format, size_bytes, error_message, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, 0), NULLIF($9, 0), NULLIF($10, ''), NULLIF($11, 0), NULLIF($12, ''), $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			result_object_key = EXCLUDED.result_object_key,
			result_url = EXCLUDED.result_url,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			format = EXCLUDED.format,
			size_bytes = EXCLUDED.size_bytes,
			error_message = EXCLUDED.error_message,
			updated_at = EXCLUDED.updated_at
	`, job.ID, string(job.Status), layoutJSON, sourcesJSON, job.IdempotencyKey, job.ResultObjectKey, job.ResultURL,
		job.Width, job.Height, string(job.Format), job.SizeBytes, job.ErrorMessage, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save composition job: %w", err)
	}
	return nil
}

func (r *JobRepository) FindByID(ctx context.Context, id string) (*domain.CompositionJob, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+jobColumns+` FROM composition_jobs WHERE id = $1`, id)
	return scanJob(row)
}

func scanJob(row pgx.Row) (*domain.CompositionJob, error) {
	var (
		job                                       domain.CompositionJob
		status, format                            string
		layoutJSON, sourcesJSON                   []byte
		idempotencyKey, resultKey, resultURL, msg *string
		width, height                             *int
		sizeBytes                                 *int64
	)

	err := row.Scan(
		&job.ID, &status, &layoutJSON, &sourcesJSON, &idempotencyKey, &resultKey, &resultURL,
		&width, &height, &format, &sizeBytes, &msg, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJobNotFound
		}
		return nil, fmt.Errorf("scan composition job: %w", err)
	}

	var layout layoutRecord
	if err := json.Unmarshal(layoutJSON, &layout); err != nil {
		return nil, fmt.Errorf("unmarshal layout: %w", err)
	}
	var sources []sourceRecord
	if err := json.Unmarshal(sourcesJSON, &sources); err != nil {
		return nil, fmt.Errorf("unmarshal sources: %w", err)
	}

	job.Status = domain.JobStatus(status)
	job.Layout = layout.toDomain()
	job.Sources = sourceRecords(sources).toDomain()
	job.Format = domain.OutputFormat(format)
	if idempotencyKey != nil {
		job.IdempotencyKey = *idempotencyKey
	}
	if resultKey != nil {
		job.ResultObjectKey = *resultKey
	}
	if resultURL != nil {
		job.ResultURL = *resultURL
	}
	if width != nil {
		job.Width = *width
	}
	if height != nil {
		job.Height = *height
	}
	if sizeBytes != nil {
		job.SizeBytes = *sizeBytes
	}
	if msg != nil {
		job.ErrorMessage = *msg
	}

	return &job, nil
}
