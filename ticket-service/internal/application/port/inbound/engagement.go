package inbound

import (
	"context"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type WishlistUseCase interface {
	Add(ctx context.Context, actor Principal, eventID int64) error
	Remove(ctx context.Context, actor Principal, eventID int64) error
	List(ctx context.Context, actor Principal, limit, offset int) ([]domain.Event, error)
}

type NotificationUseCase interface {
	List(ctx context.Context, actor Principal, unreadOnly bool, afterID int64, limit int) (domain.Page[domain.Notification], error)
	UnreadCount(ctx context.Context, actor Principal) (int, error)
	MarkRead(ctx context.Context, actor Principal, id int64) error
	MarkAllRead(ctx context.Context, actor Principal) (int, error)
	PushPending(ctx context.Context) (int, error)
	SendEmails(ctx context.Context) (int, error)
}
