package postgres

import (
	"context"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) Ping(ctx context.Context) error { return r.pool.Ping(ctx) }

func (r *Repository) Upsert(ctx context.Context, source string, videos []model.Video) error {
	if len(videos) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, v := range videos {
		tags := v.Tags
		if tags == nil {
			tags = []string{}
		}
		batch.Queue(`
			INSERT INTO videos (id, title, description, tags, status, duration, views, source, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title, description = EXCLUDED.description, tags = EXCLUDED.tags,
				status = EXCLUDED.status, duration = EXCLUDED.duration, views = EXCLUDED.views,
				source = EXCLUDED.source, updated_at = now()`,
			v.ID, v.Title, v.Description, tags, v.Status, v.Duration, v.Views, source, v.CreatedAt)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range videos {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

func (r *Repository) List(ctx context.Context) ([]model.Video, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, title, description, tags, status, duration, views, created_at FROM videos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Video
	for rows.Next() {
		var v model.Video
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.Tags, &v.Status, &v.Duration, &v.Views, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM videos WHERE id = $1`, id)
	return err
}

func (r *Repository) DeleteMissing(ctx context.Context, source string, keep []string) error {
	if keep == nil {
		keep = []string{}
	}
	_, err := r.pool.Exec(ctx, `DELETE FROM videos WHERE source = $1 AND id <> ALL($2)`, source, keep)
	return err
}

func (r *Repository) Add(ctx context.Context, in model.Interaction) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO interactions (user_id, video_id, type, value, created_at) VALUES ($1, $2, $3, $4, $5)`,
		in.UserID, in.VideoID, string(in.Type), in.Value, in.At)
	return err
}

func (r *Repository) Since(ctx context.Context, since time.Time) ([]model.Interaction, error) {
	return r.query(ctx, `SELECT user_id, video_id, type, value, created_at FROM interactions WHERE created_at >= $1`, since)
}

func (r *Repository) ByUser(ctx context.Context, userID string, limit int) ([]model.Interaction, error) {
	return r.query(ctx, `SELECT user_id, video_id, type, value, created_at FROM interactions
		WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
}

func (r *Repository) DeleteByVideo(ctx context.Context, videoID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM interactions WHERE video_id = $1`, videoID)
	return err
}

func (r *Repository) DeleteByUserVideo(ctx context.Context, userID, videoID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM interactions WHERE user_id = $1 AND video_id = $2`, userID, videoID)
	return err
}

func (r *Repository) query(ctx context.Context, sql string, args ...any) ([]model.Interaction, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Interaction
	for rows.Next() {
		var in model.Interaction
		var typ string
		if err := rows.Scan(&in.UserID, &in.VideoID, &typ, &in.Value, &in.At); err != nil {
			return nil, err
		}
		in.Type = model.InteractionType(typ)
		out = append(out, in)
	}
	return out, rows.Err()
}
