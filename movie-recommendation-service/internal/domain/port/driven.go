package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
)

type VideoRepository interface {
	Upsert(ctx context.Context, source string, videos []model.Video) error
	List(ctx context.Context) ([]model.Video, error)
	Delete(ctx context.Context, id string) error
	DeleteMissing(ctx context.Context, source string, keep []string) error
}

type InteractionRepository interface {
	Add(ctx context.Context, in model.Interaction) error
	Since(ctx context.Context, since time.Time) ([]model.Interaction, error)
	ByUser(ctx context.Context, userID string, limit int) ([]model.Interaction, error)
	DeleteByVideo(ctx context.Context, videoID string) error
	DeleteByUserVideo(ctx context.Context, userID, videoID string) error
}

type SnapshotRepository interface {
	Save(ctx context.Context, id, scope string, ranked []engine.Scored, expiresAt time.Time) error
	Load(ctx context.Context, id, scope string, now time.Time) (ranked []engine.Scored, ok bool, err error)
	DeleteExpired(ctx context.Context, now time.Time) error
}

type CatalogSource interface {
	Fetch(ctx context.Context) ([]model.Video, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
