package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
)

type MessageSender interface {
	SendMessage(ctx context.Context, msg model.OutgoingMessage) error
}

type ChannelSender interface {
	Channel() model.Channel
	SendMessage(ctx context.Context, msg model.OutgoingMessage) error
}

type ChannelPublisher interface {
	Channel() model.Channel
	Publish(ctx context.Context, post model.Post) (externalID string, err error)
}

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	Update(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	List(ctx context.Context, limit, offset int32, status model.PostStatus, campaign string) ([]model.Post, error)
	DueForPublish(ctx context.Context, before time.Time) ([]model.Post, error)
	ActiveRecurring(ctx context.Context) ([]model.Post, error)
}
