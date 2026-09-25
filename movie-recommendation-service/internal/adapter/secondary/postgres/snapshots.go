package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SnapshotStore struct{ pool *pgxpool.Pool }

func NewSnapshotStore(pool *pgxpool.Pool) *SnapshotStore { return &SnapshotStore{pool: pool} }

type storedItem struct {
	VideoID   string  `json:"v"`
	Score     float64 `json:"s"`
	Reason    string  `json:"r"`
	BecauseID string  `json:"b,omitempty"`
}

func (r *SnapshotStore) Save(ctx context.Context, id, scope string, ranked []engine.Scored, expiresAt time.Time) error {
	items := make([]storedItem, len(ranked))
	for i, s := range ranked {
		items[i] = storedItem{s.VideoID, s.Score, s.Reason, s.BecauseID}
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO ranking_snapshots (id, scope, items, expires_at) VALUES ($1, $2, $3, $4)`,
		id, scope, payload, expiresAt)
	return err
}

func (r *SnapshotStore) Load(ctx context.Context, id, scope string, now time.Time) ([]engine.Scored, bool, error) {
	var payload []byte
	err := r.pool.QueryRow(ctx,
		`SELECT items FROM ranking_snapshots WHERE id = $1 AND scope = $2 AND expires_at > $3`,
		id, scope, now).Scan(&payload)
	if err == pgx.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var items []storedItem
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, false, err
	}
	ranked := make([]engine.Scored, len(items))
	for i, it := range items {
		ranked[i] = engine.Scored{VideoID: it.VideoID, Score: it.Score, Reason: it.Reason, BecauseID: it.BecauseID}
	}
	return ranked, true, nil
}

func (r *SnapshotStore) DeleteExpired(ctx context.Context, now time.Time) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM ranking_snapshots WHERE expires_at <= $1`, now)
	return err
}
