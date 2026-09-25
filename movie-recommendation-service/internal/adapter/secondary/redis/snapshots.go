package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

const keyPrefix = "movie-rec:snapshot:"

type SnapshotStore struct{ rdb *goredis.Client }

var _ port.SnapshotRepository = (*SnapshotStore)(nil)

func NewSnapshotStore(rdb *goredis.Client) *SnapshotStore { return &SnapshotStore{rdb: rdb} }

func New(lc fx.Lifecycle, cfg *config.Config) *SnapshotStore {
	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password})
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := rdb.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("redis: %w", err)
			}
			return nil
		},
		OnStop: func(context.Context) error { return rdb.Close() },
	})
	return NewSnapshotStore(rdb)
}

type storedItem struct {
	VideoID   string  `json:"v"`
	Score     float64 `json:"s"`
	Reason    string  `json:"r"`
	BecauseID string  `json:"b,omitempty"`
}

type record struct {
	Scope string       `json:"scope"`
	Items []storedItem `json:"items"`
}

func (s *SnapshotStore) Save(ctx context.Context, id, scope string, ranked []engine.Scored, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	rec := record{Scope: scope, Items: make([]storedItem, len(ranked))}
	for i, r := range ranked {
		rec.Items[i] = storedItem{r.VideoID, r.Score, r.Reason, r.BecauseID}
	}
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, keyPrefix+id, payload, ttl).Err()
}

func (s *SnapshotStore) Load(ctx context.Context, id, scope string, _ time.Time) ([]engine.Scored, bool, error) {
	payload, err := s.rdb.Get(ctx, keyPrefix+id).Bytes()
	if err == goredis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var rec record
	if err := json.Unmarshal(payload, &rec); err != nil {
		return nil, false, err
	}
	if rec.Scope != scope {
		return nil, false, nil
	}
	ranked := make([]engine.Scored, len(rec.Items))
	for i, it := range rec.Items {
		ranked[i] = engine.Scored{VideoID: it.VideoID, Score: it.Score, Reason: it.Reason, BecauseID: it.BecauseID}
	}
	return ranked, true, nil
}

func (s *SnapshotStore) DeleteExpired(context.Context, time.Time) error { return nil }
