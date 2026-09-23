package port

import (
	"context"
	"io"

	"github.com/JIeeiroSst/video-service/internal/domain/model"
)

type UploadInput struct {
	Title       string
	Description string
	ContentType string
	Body        io.Reader
}

type Sort string

const (
	SortNewest  Sort = "newest"
	SortPopular Sort = "popular"
)

type ListQuery struct {
	Query    string
	Sort     Sort
	Page     int
	PageSize int
}

type VideoUsecase interface {
	Upload(ctx context.Context, in UploadInput) (*model.Video, error)
	List(ctx context.Context, q ListQuery) (videos []model.Video, total int, err error)
	Related(ctx context.Context, id string, limit int) ([]model.Video, error)
	Get(ctx context.Context, id string) (*model.Video, error)
	Open(ctx context.Context, id string) (*model.VideoStream, error)
	Thumbnail(ctx context.Context, id string) (*model.Asset, error)
	HLS(ctx context.Context, id, rel string) (*model.Asset, error)
	RecordView(ctx context.Context, id, viewer string) error
	Delete(ctx context.Context, id string) error
}

type TranscodeUsecase interface {
	Process(ctx context.Context, id string) error
}
