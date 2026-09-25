package port

import (
	"context"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
)

type PageQuery struct {
	Page     int
	PageSize int
	Snapshot string
}

type Page struct {
	Items    []model.Recommendation `json:"items"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Snapshot string                 `json:"snapshot"`
}

type RecommendationUsecase interface {
	ForUser(ctx context.Context, userID string, q PageQuery) (*Page, error)
	Similar(ctx context.Context, videoID string, q PageQuery) (*Page, error)
	Trending(ctx context.Context, q PageQuery) (*Page, error)
	NewReleases(ctx context.Context, q PageQuery) (*Page, error)
	ContinueWatching(ctx context.Context, userID string, q PageQuery) (*Page, error)
	History(ctx context.Context, userID string, q PageQuery) (*Page, error)
	RemoveFromHistory(ctx context.Context, userID, videoID string) error
	Home(ctx context.Context, userID string) (*model.Home, error)
	RecordEvent(ctx context.Context, in model.Interaction) error
	UpsertVideo(ctx context.Context, v model.Video) error
	DeleteVideo(ctx context.Context, id string) error
	Ready(ctx context.Context) error
}

type ModelMaintainer interface {
	Sync(ctx context.Context) error
	Train(ctx context.Context) error
	Cleanup(ctx context.Context) error
}
