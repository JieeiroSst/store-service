package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
)

type BotUsecase interface {
	HandleMessage(ctx context.Context, msg model.IncomingMessage) error
}

type CreatePostInput struct {
	SaveAsDraft bool

	Title       string
	Text        string
	Hashtags    []string
	Media       []model.Media
	MediaKind   model.MediaKind
	Channels    []model.Channel
	ScheduledAt *time.Time
	Timezone    string

	CronExpr        string
	MaxRunsPerDay   int
	MaxRunsPerMonth int

	Campaign  string
	CreatedBy string
}

type UpdatePostInput struct {
	Title           *string
	Text            *string
	Hashtags        []string
	Media           []model.Media
	MediaKind       *model.MediaKind
	Channels        []model.Channel
	ScheduledAt     *time.Time
	Timezone        *string
	CronExpr        *string
	MaxRunsPerDay   *int
	MaxRunsPerMonth *int
	Campaign        *string
}

type ContentUsecase interface {
	CreatePost(ctx context.Context, input CreatePostInput) (*model.Post, error)
	UpdatePost(ctx context.Context, id string, input UpdatePostInput) (*model.Post, error)
	GetPost(ctx context.Context, id string) (*model.Post, error)
	ListPosts(ctx context.Context, limit, offset int32, status model.PostStatus, campaign string) ([]model.Post, error)
	SubmitForReview(ctx context.Context, id string, changedBy string) (*model.Post, error)
	ApprovePost(ctx context.Context, id string, changedBy string) (*model.Post, error)
	RejectPost(ctx context.Context, id string, changedBy string, reason string) (*model.Post, error)
	PublishNow(ctx context.Context, id string, changedBy string) (*model.Post, error)
	CancelScheduledPost(ctx context.Context, id string, changedBy string) error
	PublishDuePosts(ctx context.Context) error
}
